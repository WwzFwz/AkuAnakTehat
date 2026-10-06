// Run with foundation_test.go. Uses synthetic Kafka records and restarts services.
package foundation_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const dashboard = "http://127.0.0.1:8091"
const notifier = "http://127.0.0.1:8092"
const pemda = "http://127.0.0.1:8093"

func TestEventPipeline(t *testing.T) {
	query := func(sql string) string {
		return docker(t, sql, "exec", "-T", "canonical-db", "sh", "-ec", `exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At`)
	}
	wait := func(label string, condition func() bool) {
		t.Helper()
		until := time.Now().Add(100 * time.Second)
		for time.Now().Before(until) {
			if condition() {
				return
			}
			time.Sleep(time.Second)
		}
		t.Fatalf("timed out: %s", label)
	}
	find := func(base, path, field, id string) map[string]json.RawMessage {
		t.Helper()
		response := request(t, "GET", base+path+"?after="+url.QueryEscape(id[:len(id)-1])+"&limit=200", "", nil)
		status(t, response, 200)
		page := decode[struct {
			Data []map[string]json.RawMessage `json:"data"`
		}](t, response.body)
		for _, row := range page.Data {
			var got string
			_ = json.Unmarshal(row[field], &got)
			if got == id {
				return row
			}
		}
		return nil
	}
	publish := func(key, payload string) {
		docker(t, key+"|"+payload+"\n", "exec", "-T", "kafka", "/opt/kafka/bin/kafka-console-producer.sh", "--bootstrap-server", "kafka:9092", "--topic", "bnpb.hazard-events.v1", "--property", "parse.key=true", "--property", "key.separator=|")
	}
	for _, base := range []string{dashboard, notifier} {
		status(t, request(t, "GET", base+"/health", "", nil), 200)
		status(t, request(t, "GET", base+"/ready", "", nil), 200)
	}
	wait("outbox ACK", func() bool { return query(`SELECT count(*)>0 FROM outbox WHERE published_at IS NOT NULL;`) == "t" })
	canonical := query(`SELECT hazard_id FROM hazard_events ORDER BY ingested_at LIMIT 1;`)
	wait("source snapshot in dashboard", func() bool { return find(dashboard, "/view", "hazard_id", canonical) != nil })
	t.Log("mock -> ingest -> PostgreSQL outbox -> Kafka -> dashboard verified")

	aggregatorID := docker(t, "", "ps", "-q", "aggregator")
	// A new group AND empty store prove historical replay on every run, even
	// when the regular pemda container has already consumed the test records.
	suffix := fmt.Sprintf("%012x", time.Now().UnixNano()&0xffffffffffff)
	if err := os.MkdirAll(".local", 0700); err != nil {
		t.Fatal(err)
	}
	override, err := filepath.Abs(filepath.Join(".local", "late-subscriber-"+suffix+".yml"))
	if err != nil {
		t.Fatal(err)
	}
	body := "services:\n  pemda-portal:\n    environment:\n      KAFKA_GROUP_ID: pemda-late-" + suffix + "\n      SQLITE_PATH: /tmp/pemda-late.db\n"
	if err := os.WriteFile(override, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		docker(t, "", "--profile", "demo", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "pemda-portal")
		_ = os.Remove(override)
	})
	docker(t, "", "-f", "docker-compose.yml", "-f", override, "--profile", "demo", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "pemda-portal")
	wait("late subscriber catches historical event", func() bool { return find(pemda, "/view", "hazard_id", canonical) != nil })
	if docker(t, "", "ps", "-q", "aggregator") != aggregatorID {
		t.Fatal("new subscriber replaced producer")
	}
	t.Log("fresh pemda group and empty store replayed history without replacing Aggregator")
	// Restore the regular persistent volume/group before restart/dedup checks.
	docker(t, "", "--profile", "demo", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "pemda-portal")

	hazardID := "00000000-0000-4000-8000-" + suffix
	eventID := "00000001-0000-4000-8000-" + suffix
	payload := func(version int, severity string) string {
		return fmt.Sprintf(`{"schema_version":1,"event_id":%q,"event_type":"hazard.upserted","hazard_id":%q,"version":%d,"correlation_id":%q,"published_at":"2026-10-01T00:00:01Z","hazard":{"hazard_id":%q,"source":"BMKG","source_ref_id":"EVENT-TEST","hazard_type":"SEISMIC","severity":%q,"area_name":"Synthetic event check","latitude":0,"longitude":0,"occurred_at":"2026-10-01T00:00:00Z","ingested_at":"2026-10-01T00:00:01Z","attributes":{"future_integer":9007199254740993}},"future_envelope":true}`, eventID, hazardID, version, suffix, hazardID, severity)
	}
	original := payload(2, "AWAS")
	publish(hazardID, original)
	wait("synthetic alert processed", func() bool {
		return find(dashboard, "/view", "hazard_id", hazardID) != nil && find(notifier, "/processed", "event_id", eventID) != nil && find(pemda, "/view", "hazard_id", hazardID) != nil
	})
	first := find(notifier, "/processed", "event_id", eventID)
	if string(first["alert"]) != "true" {
		t.Fatal("AWAS was not sent")
	}
	publish(hazardID, original)
	// A distinct older event may arrive on replay; views must keep version 2.
	eventID = "00000002-0000-4000-8000-" + suffix
	publish(hazardID, payload(1, "NORMAL"))
	wait("older event consumed by notifier", func() bool { return find(notifier, "/processed", "event_id", eventID) != nil })
	if string(find(notifier, "/processed", "event_id", eventID)["alert"]) != "false" {
		t.Fatal("NORMAL alert sent")
	}
	for _, base := range []string{dashboard, pemda} {
		row := find(base, "/view", "hazard_id", hazardID)
		if string(row["version"]) != "2" || string(row["future_envelope"]) != "true" || !strings.Contains(string(row["hazard"]), "9007199254740993") {
			t.Fatal("replay damaged latest view or additive fields")
		}
	}
	eventID = "00000001-0000-4000-8000-" + suffix
	if string(find(notifier, "/processed", "event_id", eventID)["processed_at"]) != string(first["processed_at"]) {
		t.Fatal("duplicate event changed durable notifier marker")
	}
	docker(t, "", "restart", "dashboard-updater", "notifier", "pemda-portal")
	docker(t, "", "--profile", "demo", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "dashboard-updater", "notifier", "pemda-portal")
	publish(hazardID, original)
	wait("SQLite survives restart", func() bool {
		return find(dashboard, "/view", "hazard_id", hazardID) != nil && find(pemda, "/view", "hazard_id", hazardID) != nil
	})
	if string(find(notifier, "/processed", "event_id", eventID)["processed_at"]) != string(first["processed_at"]) {
		t.Fatal("restart lost notifier dedup")
	}
	t.Log("replay/stale versions, precise additive fields, SIAGA/AWAS filtering and SQLite restart persistence verified")

	poison := "poison-" + suffix
	publish(poison, `{"bad":"`+poison+`"}`)
	eventID = "00000003-0000-4000-8000-" + suffix
	publish(hazardID, payload(3, "SIAGA"))
	wait("all groups continue after DLQ", func() bool {
		return string(find(dashboard, "/view", "hazard_id", hazardID)["version"]) == "3" && string(find(pemda, "/view", "hazard_id", hazardID)["version"]) == "3" && find(notifier, "/processed", "event_id", eventID) != nil
	})
	dlq := docker(t, "", "exec", "-T", "kafka", "sh", "-c", `/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server kafka:9092 --topic bnpb.hazard-events.v1.dlq --from-beginning --timeout-ms 5000 --property print.headers=true; code=$?; [ "$code" -le 1 ]`)
	for _, group := range []string{"dashboard-updater", "notifier", "pemda-portal"} {
		found := false
		for _, line := range strings.Split(dlq, "\n") {
			if strings.Contains(line, poison) && strings.Contains(line, "consumer_group:"+group) && strings.Contains(line, "failure_reason:invalid_event") && strings.Contains(line, "attempts:1") && strings.Contains(line, "source_offset:") {
				found = true
			}
		}
		if !found {
			t.Fatal("missing DLQ payload/headers for", group)
		}
	}
	t.Log("invalid record reached DLQ for all three groups; subsequent valid record processed")

	docker(t, "", "stop", "dashboard-updater")
	t.Cleanup(func() { docker(t, "", "start", "dashboard-updater") })
	eventID = "00000004-0000-4000-8000-" + suffix
	publish(hazardID, payload(4, "AWAS"))
	wait("other groups independent", func() bool {
		return string(find(pemda, "/view", "hazard_id", hazardID)["version"]) == "4" && find(notifier, "/processed", "event_id", eventID) != nil
	})
	docker(t, "", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "dashboard-updater")
	wait("offline group catches up", func() bool { return string(find(dashboard, "/view", "hazard_id", hazardID)["version"]) == "4" })
	t.Log("paused dashboard caught up; notifier and pemda continued independently")
	for _, service := range []string{"dashboard-updater", "notifier", "pemda-portal"} {
		logs := docker(t, "", "logs", "--no-log-prefix", "--tail=200", service)
		if !strings.Contains(logs, `"correlation_id":"`+suffix+`"`) || !strings.Contains(logs, `"msg":"record_completed"`) {
			t.Fatal("event trace absent after offset commit", service)
		}
	}

	checkpoint := query(`SELECT watermark FROM checkpoints WHERE endpoint='bmkg.seismic-events';`)
	docker(t, "", "stop", "kafka")
	t.Cleanup(func() { docker(t, "", "start", "kafka") })
	wait("ingest continues while broker is down", func() bool {
		return query(`SELECT watermark FROM checkpoints WHERE endpoint='bmkg.seismic-events';`) != checkpoint && query(`SELECT count(*)>0 FROM outbox WHERE published_at IS NULL;`) == "t"
	})
	for _, address := range []string{dashboard, notifier, pemda} {
		status(t, request(t, "GET", address+"/health", "", nil), 200)
		started := time.Now()
		status(t, request(t, "GET", address+"/ready", "", nil), 503)
		if time.Since(started) > 3500*time.Millisecond {
			t.Fatal("consumer readiness exceeded dependency deadline", address)
		}
	}
	boundary := query(`SELECT max(id) FROM outbox;`)
	docker(t, "", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "90", "kafka")
	wait("pending events drain after broker recovery", func() bool {
		return query(`SELECT count(*)=0 FROM outbox WHERE id<=`+boundary+` AND published_at IS NULL;`) == "t"
	})
	t.Log("Kafka outage did not block ingest; retained outbox drained after recovery")
}
