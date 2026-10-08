// Run together with foundation_test.go to reuse its HTTP/Docker helpers.
package foundation_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestIngestPipeline(t *testing.T) {
	query := func(sql string) string {
		t.Helper()
		return docker(t, sql, "exec", "-T", "canonical-db", "sh", "-ec", `exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At`)
	}
	wait := func(label string, condition func() bool) {
		t.Helper()
		deadline := time.Now().Add(50 * time.Second)
		for time.Now().Before(deadline) {
			if condition() {
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatalf("timed out: %s", label)
	}
	wait("initial ingest", func() bool {
		return query(`SELECT count(DISTINCT source)=2 FROM hazard_events;`) == "t" && query(`SELECT count(*)=3 FROM checkpoints;`) == "t"
	})
	wait("both sources healthy", func() bool { return query(`SELECT bool_and(status='HEALTHY') FROM source_status;`) == "t" })
	if query(`SELECT count(*)>0 FROM outbox;`) != "t" {
		t.Fatal("ingest must create outbox snapshots")
	}
	if query(`SELECT count(*) FROM outbox WHERE payload->>'hazard_id'<>hazard_id::text OR (payload->>'version')::bigint<>version OR payload->'hazard'->>'hazard_id'<>hazard_id::text;`) != "0" {
		t.Fatal("outbox envelope differs from row identity")
	}
	c := credentials(t)
	admin := map[string]string{"X-Admin-Key": c["PVMBG_ADMIN_KEY"], "Content-Type": "application/json"}
	initial := decode[struct {
		Simulation struct {
			Outage bool   `json:"outage"`
			Mode   string `json:"mode"`
			Schema int    `json:"schema_version"`
		} `json:"simulation"`
	}](t, request(t, "GET", pvmbg+"/health", "", nil).body)
	t.Cleanup(func() {
		status(t, request(t, "POST", pvmbg+"/admin/outage", fmt.Sprintf(`{"enabled":%t,"mode":%q}`, initial.Simulation.Outage, initial.Simulation.Mode), admin), 200)
		status(t, request(t, "POST", pvmbg+"/admin/schema-version", fmt.Sprintf(`{"version":%d}`, initial.Simulation.Schema), admin), 200)
	})
	schemaStart := query(`SELECT clock_timestamp();`)
	status(t, request(t, "POST", pvmbg+"/admin/schema-version", `{"version":2}`, admin), 200)
	wait("unknown field reaches canonical JSONB", func() bool {
		return query(`SELECT count(*)>0 FROM hazard_events WHERE source='PVMBG' AND attributes ? 'confidence_level' AND ingested_at>='`+schemaStart+`';`) == "t"
	})
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":true,"mode":"hang"}`, admin), 200)
	wait("PVMBG timeout opens breaker", func() bool {
		return query(`SELECT status='DOWN' AND consecutive_failures>=3 FROM source_status WHERE source='PVMBG';`) == "t"
	})
	held := query(`SELECT watermark FROM checkpoints WHERE endpoint='pvmbg.volcanic-reports';`)
	beforeBMKG := query(`SELECT count(*) FROM hazard_events WHERE source='BMKG';`)
	wait("BMKG continues during PVMBG outage", func() bool { return query(`SELECT count(*) FROM hazard_events WHERE source='BMKG';`) != beforeBMKG })
	if query(`SELECT watermark FROM checkpoints WHERE endpoint='pvmbg.volcanic-reports';`) != held {
		t.Fatal("failed PVMBG polling advanced checkpoint")
	}
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":false}`, admin), 200)
	wait("PVMBG automatic recovery", func() bool { return query(`SELECT status='HEALTHY' FROM source_status WHERE source='PVMBG';`) == "t" })
	if query(`SELECT watermark FROM checkpoints WHERE endpoint='pvmbg.volcanic-reports';`) == held {
		t.Fatal("recovered source did not checkpoint")
	}
	marker := query(`SELECT hazard_id FROM hazard_events ORDER BY ingested_at,hazard_id LIMIT 1;`)
	beforeRestart := query(`SELECT clock_timestamp();`)
	migrationBefore := query(`SELECT version::text || chr(58) || dirty::text FROM schema_migrations;`)
	docker(t, "", "restart", "aggregator")
	wait("Aggregator resumes from durable checkpoints", func() bool {
		return query(`SELECT max(last_attempt_at)>'`+beforeRestart+`' FROM source_status;`) == "t"
	})
	if query(`SELECT count(*) FROM hazard_events WHERE hazard_id='`+marker+`';`) != "1" {
		t.Fatal("restart lost stable hazard identity")
	}
	if query(`SELECT version::text || chr(58) || dirty::text FROM schema_migrations;`) != migrationBefore || !strings.HasSuffix(migrationBefore, ":false") {
		t.Fatal("migration version changed on restart")
	}
	logs := docker(t, "", "logs", "--no-log-prefix", "--tail=400", "aggregator")
	if !strings.Contains(logs, `"msg":"schema_drift"`) {
		t.Fatal("schema drift was not logged")
	}
	t.Log("mock -> canonical store -> pending outbox; schema drift, isolated outage, recovery and restart passed")
}
