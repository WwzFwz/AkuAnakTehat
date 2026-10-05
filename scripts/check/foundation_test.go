// Run from the repository root after docker compose up -d --build:
// go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m
// These integration tests change mock simulation state and restart infrastructure.
package foundation_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const bmkg = "http://127.0.0.1:8081"
const pvmbg = "http://127.0.0.1:8082"
const auth = "http://127.0.0.1:8090"
const api = "http://127.0.0.1:8080"

var httpClient = &http.Client{Timeout: 5 * time.Second}

func TestMain(m *testing.M) {
	// go test sets cwd to the package directory, including named-file runs.
	root, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot locate repository")
		os.Exit(1)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "docker-compose.yml")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			fmt.Fprintln(os.Stderr, "cannot locate repository")
			os.Exit(1)
		}
		root = parent
	}
	if os.Chdir(root) != nil {
		fmt.Fprintln(os.Stderr, "cannot enter repository")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

type reply struct {
	status int
	header http.Header
	body   []byte
}

func request(t *testing.T, method, address, body string, headers map[string]string) reply {
	t.Helper()
	req, err := http.NewRequest(method, address, strings.NewReader(body))
	if err != nil {
		t.Fatal("invalid test request")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s failed: %v", method, address, err)
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	if err != nil {
		t.Fatal("reading response failed")
	}
	return reply{res.StatusCode, res.Header, b}
}

func status(t *testing.T, r reply, want int) {
	t.Helper()
	// Never print bodies: token responses contain credentials.
	if r.status != want {
		t.Fatalf("HTTP status = %d; want %d (body withheld)", r.status, want)
	}
}

func decode[T any](t *testing.T, b []byte) T {
	t.Helper()
	var v T
	if json.Unmarshal(b, &v) != nil {
		t.Fatal("invalid response JSON (body withheld)")
	}
	return v
}

func credentials(t *testing.T) map[string]string {
	t.Helper()
	b, err := os.ReadFile("env/demo.env")
	if err != nil {
		t.Fatal("run from repository root after secret bootstrap")
	}
	v := map[string]string{}
	for _, line := range strings.Split(string(b), "\n") {
		k, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if ok && !strings.HasPrefix(k, "#") {
			v[k] = value
		}
	}
	for _, key := range []string{"BMKG_API_KEY", "PVMBG_TOKEN", "PVMBG_ADMIN_KEY"} {
		if v[key] == "" {
			t.Fatalf("missing demo credential %s", key)
		}
	}
	return v
}

func waitHTTP(t *testing.T, address string, want int) {
	t.Helper()
	until := time.Now().Add(90 * time.Second)
	for time.Now().Before(until) {
		r, err := httpClient.Get(address)
		if err == nil {
			r.Body.Close()
			if r.StatusCode == want {
				return
			}
		}
		time.Sleep(time.Second)
	}
	t.Fatalf("%s did not reach HTTP %d within 90s", address, want)
}

func TestHealth(t *testing.T) {
	for _, base := range []string{bmkg, pvmbg, auth, api} {
		waitHTTP(t, base+"/health", 200)
	}
	waitHTTP(t, auth+"/ready", 200)
	// Aggregator query readiness is now part of the full Compose stack.
	status(t, request(t, "GET", api+"/ready", "", nil), 200)
}

func TestMockContracts(t *testing.T) {
	c := credentials(t)
	bh := map[string]string{"X-BMKG-Key": c["BMKG_API_KEY"], "X-Correlation-ID": "foundation-smoke"}
	ph := map[string]string{"Authorization": "Bearer " + c["PVMBG_TOKEN"]}
	for _, check := range []struct {
		address string
		headers map[string]string
	}{
		{bmkg + "/seismic-events", nil},
		{bmkg + "/seismic-events", map[string]string{"X-BMKG-Key": c["PVMBG_TOKEN"]}},
		{pvmbg + "/volcanic-reports", map[string]string{"Authorization": "Bearer " + c["BMKG_API_KEY"]}},
		{pvmbg + "/admin/outage", map[string]string{"X-Admin-Key": c["PVMBG_TOKEN"]}},
	} {
		method := "GET"
		if strings.Contains(check.address, "/admin/") {
			method = "POST"
		}
		status(t, request(t, method, check.address, "", check.headers), 401)
	}
	for _, check := range []struct {
		base, path, id, stamp string
		headers               map[string]string
		minimum               int
	}{
		{bmkg, "/seismic-events", "event_id", "occurred_at", bh, 20},
		{pvmbg, "/volcanic-reports", "report_id", "reported_at", ph, 20},
	} {
		r := request(t, "GET", check.base+check.path, "", check.headers)
		status(t, r, 200)
		rows := decode[[]map[string]any](t, r.body)
		if len(rows) < check.minimum {
			t.Fatal("seed record count too small")
		}
		seen := map[string]bool{}
		for _, row := range rows {
			id, ok := row[check.id].(string)
			if !ok || id == "" || seen[id] {
				t.Fatal("missing or duplicate source ID")
			}
			seen[id] = true
		}
		stamp, ok := rows[0][check.stamp].(string)
		if !ok {
			t.Fatal("timestamp missing")
		}
		filtered := request(t, "GET", check.base+check.path+"?since="+url.QueryEscape(stamp), "", check.headers)
		status(t, filtered, 200)
		found := false
		for _, row := range decode[[]map[string]any](t, filtered.body) {
			if row[check.id] == rows[0][check.id] {
				found = true
			}
		}
		if !found {
			t.Fatal("since must include the boundary record")
		}
		status(t, request(t, "GET", check.base+check.path+"?since=invalid", "", check.headers), 400)
		empty := request(t, "GET", check.base+check.path+"?since=2999-01-01T00:00:00Z", "", check.headers)
		status(t, empty, 200)
		if strings.TrimSpace(string(empty.body)) != "[]" {
			t.Fatal("future since must return []")
		}
	}
	r := request(t, "GET", bmkg+"/tsunami-warnings", "", bh)
	status(t, r, 200)
	if r.header.Get("X-Correlation-ID") != "foundation-smoke" {
		t.Fatal("correlation ID not propagated")
	}
	for _, row := range decode[[]map[string]any](t, r.body) {
		if _, ok := row["internal_modified_at"]; ok {
			t.Fatal("internal warning watermark leaked")
		}
	}
}

func TestPVMBGSchemaAndOutage(t *testing.T) {
	c := credentials(t)
	admin := map[string]string{"X-Admin-Key": c["PVMBG_ADMIN_KEY"], "Content-Type": "application/json"}
	ph := map[string]string{"Authorization": "Bearer " + c["PVMBG_TOKEN"]}
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
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":false}`, admin), 200)
	status(t, request(t, "POST", pvmbg+"/admin/schema-version", `{"version":1,"enabled":true}`, admin), 400)
	status(t, request(t, "POST", pvmbg+"/admin/schema-version", `{"version":2}`, admin), 200)
	// Observe the real 10s generator, without replacing its production cadence.
	deadline := time.Now().Add(20 * time.Second)
	for {
		r := request(t, "GET", pvmbg+"/volcanic-reports", "", ph)
		status(t, r, 200)
		old, newRecord := false, false
		for _, row := range decode[[]map[string]any](t, r.body) {
			if v, ok := row["confidence_level"]; ok {
				n, valid := v.(float64)
				if !valid || n < 0 || n > 1 {
					t.Fatal("invalid confidence_level")
				}
				newRecord = true
			} else {
				old = true
			}
		}
		if old && newRecord {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("schema v1/v2 records did not coexist")
		}
		time.Sleep(time.Second)
	}
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":true,"mode":"error"}`, admin), 200)
	status(t, request(t, "GET", pvmbg+"/volcanic-reports", "", ph), 503)
	status(t, request(t, "GET", pvmbg+"/health", "", nil), 200)
	status(t, request(t, "GET", pvmbg+"/ready", "", nil), 503)
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":true,"mode":"hang"}`, admin), 200)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", pvmbg+"/volcanic-reports", nil)
	req.Header.Set("Authorization", ph["Authorization"])
	r, err := httpClient.Do(req)
	if err == nil {
		r.Body.Close()
		t.Fatal("hang request should hit client deadline")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Fatal("hang request failed before deadline")
	}
	status(t, request(t, "POST", pvmbg+"/admin/outage", `{"enabled":false}`, admin), 200)
	status(t, request(t, "GET", pvmbg+"/volcanic-reports", "", ph), 200)
}

type pair struct {
	Access  string `json:"access_token"`
	Refresh string `json:"refresh_token"`
	Scope   string `json:"scope"`
	Expires int    `json:"expires_in"`
}

func grant(t *testing.T, v url.Values) reply {
	t.Helper()
	return request(t, "POST", auth+"/oauth/token", v.Encode(), map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
}
func login(t *testing.T, id string) pair {
	t.Helper()
	b, err := os.ReadFile("env/demo-clients.json")
	if err != nil {
		t.Fatal("cannot read local demo clients")
	}
	clients := decode[[]struct {
		ID     string `json:"client_id"`
		Secret string `json:"client_secret"`
	}](t, b)
	for _, c := range clients {
		if c.ID == id {
			r := grant(t, url.Values{"grant_type": {"client_credentials"}, "client_id": {id}, "client_secret": {c.Secret}})
			status(t, r, 200)
			if r.header.Get("Cache-Control") != "no-store" || r.header.Get("Pragma") != "no-cache" {
				t.Fatal("token cache headers missing")
			}
			p := decode[pair](t, r.body)
			if p.Access == "" || p.Refresh == "" || p.Expires != 60 {
				t.Fatal("invalid token response")
			}
			return p
		}
	}
	t.Fatal("missing demo client")
	return pair{}
}
func refresh(t *testing.T, token string) reply {
	t.Helper()
	return grant(t, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {token}})
}
func bearer(p pair) map[string]string {
	return map[string]string{"Authorization": "Bearer " + p.Access}
}

func TestTokenRotationAndAuthorization(t *testing.T) {
	status(t, grant(t, url.Values{"grant_type": {"client_credentials"}, "client_id": {"media"}, "client_secret": {"wrong"}}), 401)
	status(t, request(t, "GET", api+"/v1/hazards", "", nil), 401)
	p := login(t, "media")
	if p.Scope != "hazard:read:summary" {
		t.Fatal("unexpected Media scopes")
	}
	for _, path := range []string{"/v1/hazards?fields=attributes", "/v1/hazards?include=raw", "/v1/hazards/demo/raw"} {
		status(t, request(t, "GET", api+path, "", bearer(p)), 403)
	}
	status(t, request(t, "GET", api+"/v1/hazards", "", bearer(p)), 200)
	for _, id := range []string{"field-team", "bnpb-ops"} {
		privileged := login(t, id)
		status(t, request(t, "GET", api+"/v1/hazards?include=raw", "", bearer(privileged)), 200)
	}
	parts := strings.Split(p.Access, ".")
	parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"attacker"}`))
	status(t, request(t, "GET", api+"/v1/hazards", "", map[string]string{"Authorization": "Bearer " + strings.Join(parts, ".")}), 401)
	rotated := refresh(t, p.Refresh)
	status(t, rotated, 200)
	next := decode[pair](t, rotated.body)
	if next.Refresh == p.Refresh {
		t.Fatal("refresh token was not rotated")
	}
	status(t, refresh(t, p.Refresh), 400)
	status(t, refresh(t, next.Refresh), 400)
	// Two HTTP requests race on the same token: exactly one succeeds;
	// detection of the second use revokes even the winning descendant.
	race := login(t, "media")
	results := make(chan reply, 2)
	for range 2 {
		go func() {
			response, err := httpClient.PostForm(auth+"/oauth/token", url.Values{"grant_type": {"refresh_token"}, "refresh_token": {race.Refresh}})
			if err != nil {
				results <- reply{status: 0}
				return
			}
			defer response.Body.Close()
			body, err := io.ReadAll(io.LimitReader(response.Body, 8192))
			if err != nil {
				results <- reply{status: 0}
				return
			}
			results <- reply{status: response.StatusCode, body: body}
		}()
	}
	first, second := <-results, <-results
	if first.status == 400 {
		first, second = second, first
	}
	status(t, first, 200)
	status(t, second, 400)
	status(t, refresh(t, decode[pair](t, first.body).Refresh), 400)
}

func TestJWTClaimValidation(t *testing.T) {
	b, err := os.ReadFile("env/keys/jwt-private.pem")
	if err != nil {
		t.Fatal("cannot read test signing key")
	}
	block, _ := pem.Decode(b)
	if block == nil {
		t.Fatal("invalid test signing key")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		t.Fatal("invalid test signing key")
	}
	priv, ok := key.(ed25519.PrivateKey)
	if !ok {
		t.Fatal("invalid key type")
	}
	for _, field := range []string{"exp", "aud", "iss", "iat"} {
		t.Run(field, func(t *testing.T) {
			claims := map[string]any{"iss": "bnpb-auth", "aud": "bnpb-api", "sub": "media", "jti": "foundation-test", "scope": "hazard:read:summary", "iat": time.Now().Add(-time.Minute).Unix(), "exp": time.Now().Add(time.Minute).Unix()}
			switch field {
			case "exp":
				// Avoid a one-second boundary across the host and Docker VM clocks.
				claims[field] = time.Now().Add(-5 * time.Minute).Unix()
			case "iat":
				claims[field] = time.Now().Add(time.Minute).Unix()
			default:
				claims[field] = "wrong"
			}
			payload, _ := json.Marshal(claims)
			unsigned := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"EdDSA","typ":"JWT"}`)) + "." + base64.RawURLEncoding.EncodeToString(payload)
			token := unsigned + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(priv, []byte(unsigned)))
			status(t, request(t, "GET", api+"/v1/hazards", "", map[string]string{"Authorization": "Bearer " + token}), 401)
		})
	}
}

func docker(t *testing.T, input string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", append([]string{"compose"}, args...)...)
	cmd.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("docker compose %s failed: %v (output withheld)", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(stdout.String())
}

func TestPersistenceAndDependencyRecovery(t *testing.T) {
	unique := fmt.Sprintf("foundation_smoke_%d", time.Now().UnixNano())
	topic := strings.ReplaceAll(unique, "_", "-")
	psql := []string{"exec", "-T", "canonical-db", "sh", "-ec", `exec psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB" -At`}
	docker(t, "CREATE TABLE "+unique+" (value text); INSERT INTO "+unique+" VALUES ('persisted');", psql...)
	t.Cleanup(func() { docker(t, "DROP TABLE IF EXISTS "+unique+";", psql...) })
	docker(t, "", "exec", "-T", "kafka", "/opt/kafka/bin/kafka-topics.sh", "--bootstrap-server", "kafka:9092", "--create", "--topic", topic, "--partitions", "1", "--replication-factor", "1")
	t.Cleanup(func() {
		docker(t, "", "exec", "-T", "kafka", "/opt/kafka/bin/kafka-topics.sh", "--bootstrap-server", "kafka:9092", "--delete", "--topic", topic)
	})
	docker(t, "persisted\n", "exec", "-T", "kafka", "/opt/kafka/bin/kafka-console-producer.sh", "--bootstrap-server", "kafka:9092", "--topic", topic)
	session := login(t, "media")
	t.Cleanup(func() { docker(t, "", "start", "auth-store", "auth-service", "canonical-db", "kafka") })
	docker(t, "", "stop", "auth-store")
	status(t, request(t, "GET", auth+"/health", "", nil), 200)
	status(t, request(t, "GET", auth+"/ready", "", nil), 503)
	status(t, refresh(t, session.Refresh), 503)
	docker(t, "", "start", "auth-store")
	docker(t, "", "restart", "canonical-db", "kafka", "auth-service")
	waitHTTP(t, auth+"/ready", 200)
	// Redis AOF preserves the unused refresh token through its stop/start.
	status(t, refresh(t, session.Refresh), 200)
	if got := docker(t, "SELECT value FROM "+unique+";", psql...); got != "persisted" {
		t.Fatal("PostgreSQL marker missing after restart")
	}
	got := docker(t, "", "exec", "-T", "kafka", "/opt/kafka/bin/kafka-console-consumer.sh", "--bootstrap-server", "kafka:9092", "--topic", topic, "--partition", "0", "--offset", "earliest", "--max-messages", "1", "--timeout-ms", "60000")
	if got != "persisted" {
		t.Fatal("Kafka marker missing after restart")
	}
}

func TestStructuredLogsDoNotExposeCredentials(t *testing.T) {
	c := credentials(t)
	p := login(t, "media")
	status(t, request(t, "GET", bmkg+"/seismic-events", "", map[string]string{"X-BMKG-Key": c["BMKG_API_KEY"], "X-Correlation-ID": "foundation-log-check"}), 200)
	logs := docker(t, "", "logs", "--no-log-prefix", "--tail=200", "bmkg-mock", "pvmbg-mock", "auth-service", "client-api")
	for _, secret := range append([]string{p.Access, p.Refresh}, c["BMKG_API_KEY"], c["PVMBG_TOKEN"], c["PVMBG_ADMIN_KEY"]) {
		if strings.Contains(logs, secret) {
			t.Fatal("credential appeared in application logs (value withheld)")
		}
	}
	found := false
	for _, line := range strings.Split(logs, "\n") {
		var row map[string]any
		if json.Unmarshal([]byte(line), &row) == nil && row["correlation_id"] == "foundation-log-check" {
			if row["method"] != "GET" || row["status"] != float64(200) || row["latency_ms"] == nil {
				t.Fatal("request log lacks required fields")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("correlated JSON request log not found")
	}
}
