package foundation_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Administrative fixtures exercise the real relay, broker, consumers and HTTP
// path. Mapping, exact admission limits and record transactions are tested with
// isolated PostgreSQL schemas in robustness_integration_test.go.
func TestLargeEventDeliveryAndPermanentRejection(t *testing.T) {
	query := func(sql string) string {
		return docker(t, sql, "exec", "-T", "canonical-db", "sh", "-ec", `exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At`)
	}
	suffix := fmt.Sprintf("%012x", time.Now().UnixNano()&0xffffffffffff)
	ids := []string{"e0000000-0000-4000-8001-" + suffix, "e0000000-0000-4000-8002-" + suffix, "e0000000-0000-4000-8003-" + suffix}
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }
	t.Cleanup(func() {
		for _, id := range ids {
			query("BEGIN; DELETE FROM outbox WHERE hazard_id=" + quote(id) + "; DELETE FROM hazard_events WHERE hazard_id=" + quote(id) + "; COMMIT;")
		}
	})
	blobLengths := map[string]int{}
	for i, id := range ids {
		h := map[string]any{"hazard_id": id, "source": "BMKG", "source_ref_id": "large-fixture-" + id, "hazard_type": "SEISMIC", "severity": "NORMAL", "area_name": "Synthetic large fixture", "latitude": 0, "longitude": 0, "occurred_at": "2020-01-01T00:00:00Z", "ingested_at": "2020-01-01T00:00:01Z", "attributes": map[string]any{"blob": ""}}
		e := map[string]any{"schema_version": 1, "event_id": id, "event_type": "hazard.upserted", "hazard_id": id, "version": 1, "correlation_id": "large-" + suffix, "published_at": "2020-01-01T00:00:01Z", "hazard": h}
		base, _ := json.Marshal(e)
		target := 4 << 20
		if i == 1 {
			target++
		}
		if i == 2 {
			target = 2048
		}
		blobLengths[id] = target - len(base)
		h["attributes"] = map[string]any{"blob": strings.Repeat("x", blobLengths[id])}
		payload, _ := json.Marshal(e)
		attrs, _ := json.Marshal(h["attributes"])
		sql := fmt.Sprintf(`BEGIN; INSERT INTO hazard_events(hazard_id,source,source_ref_id,hazard_type,severity,area_name,latitude,longitude,occurred_at,ingested_at,attributes,version,content_hash,updated_at,last_seen_at) VALUES(%s,'BMKG',%s,'SEISMIC','NORMAL','Synthetic large fixture',0,0,'2020-01-01','2020-01-01',%s,1,decode('00','hex'),now(),now()); INSERT INTO outbox(event_id,hazard_id,version,payload,created_at) VALUES(%s,%s,1,%s,now()); COMMIT;`, quote(id), quote("large-fixture-"+id), quote(string(attrs)), quote(id), quote(id), quote(string(payload)))
		query(sql)
		t.Logf("fixture %d compact envelope bytes=%d", i, len(payload))
	}
	deadline := time.Now().Add(90 * time.Second)
	for {
		good := query(fmt.Sprintf(`SELECT count(*)=2 FROM outbox WHERE hazard_id IN (%s,%s) AND published_at IS NOT NULL AND rejected_at IS NULL;`, quote(ids[0]), quote(ids[2]))) == "t"
		rejected := query(fmt.Sprintf(`SELECT count(*)=1 FROM outbox WHERE hazard_id=%s AND rejected_at IS NOT NULL AND published_at IS NULL AND rejection_reason='event_size_limit';`, quote(ids[1]))) == "t"
		if good && rejected {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("relay did not isolate permanent failure")
		}
		time.Sleep(time.Second)
	}
	for _, base := range []string{"http://127.0.0.1:8091", "http://127.0.0.1:8093", "http://127.0.0.1:8092"} {
		for _, id := range []string{ids[0], ids[2]} {
			path, field := "/view", "hazard_id"
			if strings.HasSuffix(base, ":8092") {
				path, field = "/processed", "event_id"
			}
			for {
				response := request(t, "GET", base+path+"?limit=2&after="+url.QueryEscape(id[:len(id)-1]), "", nil)
				status(t, response, 200)
				page := decode[struct {
					Data []map[string]json.RawMessage `json:"data"`
				}](t, response.body)
				found := false
				for _, row := range page.Data {
					var got string
					_ = json.Unmarshal(row[field], &got)
					if got == id {
						found = true
						if path == "/view" {
							var h struct {
								Attributes map[string]string `json:"attributes"`
							}
							if json.Unmarshal(row["hazard"], &h) != nil || len(h.Attributes["blob"]) != blobLengths[id] {
								t.Fatal("consumer truncated large event")
							}
						}
					}
				}
				if found {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("consumer failed large/following event", base)
				}
				time.Sleep(time.Second)
			}
		}
	}
	raw := request(t, "GET", api+"/v1/hazards/"+ids[0]+"/raw", "", bearer(login(t, "field-team")))
	status(t, raw, 200)
	var h struct {
		Attributes map[string]string `json:"attributes"`
		Sources    []any             `json:"sources"`
	}
	if json.Unmarshal(raw.body, &h) != nil || len(h.Attributes["blob"]) != blobLengths[ids[0]] || len(h.Sources) != 1 {
		t.Fatal("raw HTTP path truncated event or freshness")
	}
	media := request(t, "GET", api+"/v1/hazards/"+ids[0], "", bearer(login(t, "media")))
	status(t, media, 200)
	summary := decode[map[string]json.RawMessage](t, media.body)
	if summary["attributes"] != nil || summary["source_ref_id"] != nil || summary["sources"] == nil {
		t.Fatal("Media projection or freshness failed")
	}
	t.Log("4 MiB event delivered intact to all consumers and raw API; 4 MiB + 1 retained as rejected; following event delivered; Media remains restricted")
}
