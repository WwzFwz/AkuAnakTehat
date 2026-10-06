// Run with foundation_test.go after starting the complete Compose stack.
package foundation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type hazardPage struct {
	Data    []map[string]json.RawMessage `json:"data"`
	Cursor  string                       `json:"next_cursor"`
	Sources []struct {
		Source, Status string
		StaleSince     string `json:"stale_since"`
	} `json:"sources"`
}

func readPage(t *testing.T, p pair, path string) hazardPage {
	t.Helper()
	r := request(t, "GET", api+path, "", bearer(p))
	status(t, r, 200)
	return decode[hazardPage](t, r.body)
}
func waitQuery(t *testing.T, label string, condition func() bool) {
	t.Helper()
	until := time.Now().Add(45 * time.Second)
	for time.Now().Before(until) {
		if condition() {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatal("timed out:", label)
}
func rawString(raw json.RawMessage) string { var s string; _ = json.Unmarshal(raw, &s); return s }

func TestQueryIntegration(t *testing.T) {
	media, field := login(t, "media"), login(t, "field-team")
	status(t, request(t, "GET", api+"/ready", "", nil), 200)
	page := readPage(t, media, "/v1/hazards?limit=2")
	if len(page.Data) != 2 || page.Cursor == "" || len(page.Sources) != 2 {
		t.Fatal("missing data, cursor or sources")
	}
	allow := map[string]bool{"hazard_id": true, "source": true, "hazard_type": true, "severity": true, "area_name": true, "occurred_at": true, "ingested_at": true}
	for _, row := range page.Data {
		for key := range row {
			if !allow[key] {
				t.Fatal("raw field leaked to Media", key)
			}
		}
	}
	next := readPage(t, media, "/v1/hazards?limit=2&cursor="+url.QueryEscape(page.Cursor))
	for _, a := range page.Data {
		for _, b := range next.Data {
			if string(a["hazard_id"]) == string(b["hazard_id"]) {
				t.Fatal("pagination repeated an ID")
			}
		}
	}
	id := rawString(page.Data[0]["hazard_id"])
	for _, identity := range []pair{field, login(t, "bnpb-ops")} {
		r := request(t, "GET", api+"/v1/hazards/"+id+"/raw", "", bearer(identity))
		status(t, r, 200)
		row := decode[map[string]json.RawMessage](t, r.body)
		for _, key := range []string{"attributes", "latitude", "longitude", "source_ref_id"} {
			if row[key] == nil {
				t.Fatal("raw field absent", key)
			}
		}
	}
	for _, path := range []string{"/v1/hazards?include=raw", "/v1/hazards?fields=hazard_id,latitude", "/v1/hazards/" + id + "/raw"} {
		status(t, request(t, "GET", api+path, "", bearer(media)), 403)
	}
	selected := readPage(t, media, "/v1/hazards?fields=hazard_id,severity&limit=2")
	for _, row := range selected.Data {
		if len(row) != 2 {
			t.Fatal("field selection ignored")
		}
	}
	for _, path := range []string{"/v1/hazards?limit=0", "/v1/hazards?limit=501", "/v1/hazards?cursor=bad", "/v1/hazards?severity=OTHER", "/v1/hazards/seismic?type=VOLCANIC", "/v1/hazards?since=bad", "/v1/hazards?type=SEISMIC&type=VOLCANIC"} {
		status(t, request(t, "GET", api+path, "", bearer(field)), 400)
	}
	empty := readPage(t, field, "/v1/hazards?since=2099-01-01T00:00:00Z")
	if empty.Data == nil || len(empty.Data) != 0 {
		t.Fatal("empty filter contract broken")
	}
	status(t, request(t, "GET", api+"/v1/hazards/00000000-0000-4000-8000-000000000000", "", bearer(field)), 404)
	trace := "query-integration-" + fmt.Sprint(time.Now().UnixNano())
	headers := bearer(field)
	headers["X-Correlation-ID"] = trace
	r := request(t, "GET", api+"/v1/hazards/seismic?limit=2", "", headers)
	status(t, r, 200)
	if r.header.Get("X-Correlation-ID") != trace {
		t.Fatal("correlation response lost")
	}
	for _, service := range []string{"client-api", "aggregator"} {
		logs := docker(t, "", "logs", "--no-log-prefix", "--tail=100", service)
		if !strings.Contains(logs, trace) || !strings.Contains(logs, "latency_ms") {
			t.Fatal("trace/latency absent", service)
		}
		for _, line := range strings.Split(logs, "\n") {
			if strings.Contains(line, trace) {
				t.Log(service, line)
			}
		}
	}
	t.Log("P3: Media allowlist, raw rejection, distinct privileged identities; query filters, pagination, errors and cross-service trace passed")
}

func TestDynamicSchemaAndStaleAPI(t *testing.T) {
	credentials := credentials(t)
	admin := map[string]string{"X-Admin-Key": credentials["PVMBG_ADMIN_KEY"], "Content-Type": "application/json"}
	original := decode[struct {
		Simulation struct {
			Schema int `json:"schema_version"`
			Outage bool
			Mode   string
		}
	}](t, request(t, "GET", pvmbg+"/health", "", nil).body)
	t.Cleanup(func() {
		status(t, request(t, "POST", pvmbg+"/admin/outage", fmt.Sprintf(`{"enabled":%t,"mode":%q}`, original.Simulation.Outage, original.Simulation.Mode), admin), 200)
		status(t, request(t, "POST", pvmbg+"/admin/schema-version", fmt.Sprintf(`{"version":%d}`, original.Simulation.Schema), admin), 200)
	})
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":false}`, admin), 200)
	status(t, request(t, "POST", pvmbg+"/admin/schema-version", `{"version":1}`, admin), 200)
	field := login(t, "field-team")
	var oldID string
	waitQuery(t, "schema 1 record", func() bool {
		p := readPage(t, field, "/v1/hazards/volcanic?limit=100&include=raw")
		for _, h := range p.Data {
			var a map[string]json.RawMessage
			_ = json.Unmarshal(h["attributes"], &a)
			if a["confidence_level"] == nil {
				oldID = rawString(h["hazard_id"])
				return true
			}
		}
		return false
	})
	// Let this record leave the overlap window so schema-v2 polling does not rewrite it.
	time.Sleep(15 * time.Second)
	before := docker(t, "", "ps", "-q", "aggregator", "pvmbg-mock", "client-api")
	migration := docker(t, `SELECT version FROM schema_migrations;`, "exec", "-T", "canonical-db", "sh", "-ec", `exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At`)
	status(t, request(t, "POST", pvmbg+"/admin/schema-version", `{"version":2}`, admin), 200)
	waitQuery(t, "confidence_level readable through API", func() bool {
		p := readPage(t, field, "/v1/hazards/volcanic?limit=100&include=raw")
		for _, h := range p.Data {
			var a map[string]json.RawMessage
			_ = json.Unmarshal(h["attributes"], &a)
			if a["confidence_level"] != nil {
				return true
			}
		}
		return false
	})
	old := request(t, "GET", api+"/v1/hazards/"+oldID+"/raw", "", bearer(field))
	status(t, old, 200)
	h := decode[map[string]json.RawMessage](t, old.body)
	var attrs map[string]json.RawMessage
	_ = json.Unmarshal(h["attributes"], &attrs)
	if attrs["confidence_level"] != nil {
		t.Fatal("test old record was replaced by new schema")
	}
	if before != docker(t, "", "ps", "-q", "aggregator", "pvmbg-mock", "client-api") {
		t.Fatal("schema evolution replaced service")
	}
	if migration != docker(t, `SELECT version FROM schema_migrations;`, "exec", "-T", "canonical-db", "sh", "-ec", `exec psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At`) {
		t.Fatal("schema evolution required migration")
	}
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":true,"mode":"error"}`, admin), 200)
	waitQuery(t, "vulkanik stale marker", func() bool {
		p := readPage(t, field, "/v1/hazards/volcanic?limit=2")
		return len(p.Data) > 0 && len(p.Sources) == 1 && p.Sources[0].Status != "HEALTHY" && p.Sources[0].StaleSince != ""
	})
	seismic := readPage(t, field, "/v1/hazards/seismic?limit=2")
	if len(seismic.Data) == 0 || len(seismic.Sources) != 1 || seismic.Sources[0].Status != "HEALTHY" {
		t.Fatal("PVMBG outage affected seismic reads")
	}
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":false}`, admin), 200)
	field = login(t, "field-team")
	waitQuery(t, "source recovers without restart", func() bool {
		p := readPage(t, field, "/v1/hazards/volcanic?limit=2")
		return len(p.Sources) == 1 && p.Sources[0].Status == "HEALTHY" && p.Sources[0].StaleSince == ""
	})
	t.Log("P1/P4: old/new JSONB attributes coexist and remain readable without migration; P2: stale volcanic data and healthy seismic reads recover without restart")
}

func TestNaturalExpiryAndFieldCLI(t *testing.T) {
	root, _ := os.Getwd()
	binary := filepath.Join(root, ".local", "field-cli.exe")
	build := exec.Command("go", "build", "-o", binary, "./cmd/field-cli")
	build.Dir = filepath.Join(root, "tools", "field-cli")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("CLI build failed: %v %s", err, output)
	}
	b, err := os.ReadFile("env/demo-clients.json")
	if err != nil {
		t.Fatal(err)
	}
	clients := decode[[]struct {
		ID     string `json:"client_id"`
		Secret string `json:"client_secret"`
	}](t, b)
	secret := ""
	for _, c := range clients {
		if c.ID == "field-team" {
			secret = c.Secret
		}
	}
	if secret == "" {
		t.Fatal("missing field credentials")
	}
	token := login(t, "field-team")
	ctx, cancel := context.WithTimeout(context.Background(), 85*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "list", "--raw", "--limit", "1", "--watch", "65s", "--count", "2")
	cmd.Env = append(os.Environ(), "FIELD_CLI_CLIENT_ID=field-team", "FIELD_CLI_CLIENT_SECRET="+secret)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		t.Fatal("CLI session across natural expiry failed; output withheld")
	}
	decoder := json.NewDecoder(&stdout)
	for i := 0; i < 2; i++ {
		var page hazardPage
		if decoder.Decode(&page) != nil || len(page.Data) != 1 || page.Data[0]["attributes"] == nil {
			t.Fatal("CLI session returned invalid raw page")
		}
	}
	status(t, request(t, "GET", api+"/v1/hazards", "", bearer(token)), 401)
	renewed := refresh(t, token.Refresh)
	status(t, renewed, 200)
	next := decode[pair](t, renewed.body)
	status(t, request(t, "GET", api+"/v1/hazards", "", bearer(next)), 200)
	status(t, request(t, "GET", api+"/v1/hazards", "", bearer(token)), 401)
	t.Log("P3: real 60s access TTL elapsed; CLI retained its session across 65s and refreshed automatically; expired token rejected before/after refresh")
}

func TestIndependentRebuild(t *testing.T) {
	starts := map[string]string{}
	started := func(service string) string {
		t.Helper()
		id := docker(t, "", "ps", "-q", service)
		cmd := exec.Command("docker", "inspect", "--format", "{{.State.StartedAt}}", id)
		out, err := cmd.Output()
		if err != nil {
			t.Fatal("inspect service start failed")
		}
		return string(out)
	}
	for _, service := range []string{"aggregator", "client-api", "auth-service", "bmkg-mock", "pvmbg-mock", "dashboard-updater", "kafka"} {
		starts[service] = started(service)
	}
	docker(t, "", "stop", "notifier")
	t.Cleanup(func() { docker(t, "", "start", "notifier") })
	readPage(t, login(t, "media"), "/v1/hazards?limit=2")
	docker(t, "", "build", "notifier")
	docker(t, "", "up", "-d", "--no-deps", "--wait", "--wait-timeout", "60", "notifier")
	for service, stamp := range starts {
		if started(service) != stamp {
			t.Fatal("unrelated container restarted", service)
		}
	}
	t.Log("P4: notifier stopped, rebuilt and started independently; public reads continued and other container start timestamps remained unchanged")
}
