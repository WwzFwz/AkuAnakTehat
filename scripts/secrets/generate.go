// Command generate bootstraps local development secrets without rotating existing files.
package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type client struct {
	ID     string   `json:"client_id"`
	Hash   string   `json:"secret_hash"`
	Scopes []string `json:"scopes"`
}

type demoClient struct {
	ID     string   `json:"client_id"`
	Secret string   `json:"client_secret"`
	Scopes []string `json:"scopes"`
}

var managed = []string{
	"env/postgres.env", "env/redis.env", "env/bmkg-mock.env", "env/pvmbg-mock.env",
	"env/auth-service.env", "env/client-api.env", "env/aggregator.env", "env/demo.env",
	"env/clients.json", "env/demo-clients.json", "env/keys/jwt-private.pem", "env/keys/jwt-public.pem",
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	created, err := generate(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Secret bootstrap:", err)
		os.Exit(1)
	}
	if created {
		fmt.Println("Local env and Ed25519 keys created; secret values were not printed.")
	} else {
		fmt.Println("Existing local env and keys retained; no secrets rotated.")
	}
}

func generate(root string) (bool, error) {
	if _, err := os.Stat(filepath.Join(root, "docker-compose.yml")); err != nil {
		return false, errors.New("run from repository root or pass -root")
	}
	if err := os.MkdirAll(filepath.Join(root, "env", "keys"), 0700); err != nil {
		return false, err
	}
	lockPath := filepath.Join(root, "env", ".bootstrap.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return false, errors.New("bootstrap lock exists or cannot be created; check another generator is not running before removing env/.bootstrap.lock")
	}
	lock.Close()
	defer os.Remove(lockPath)
	present := 0
	for _, name := range managed {
		info, err := os.Stat(filepath.Join(root, name))
		if err == nil {
			if info.IsDir() || info.Size() == 0 {
				return false, fmt.Errorf("invalid existing file %s; repair configuration explicitly", name)
			}
			present++
		} else if !os.IsNotExist(err) {
			return false, err
		}
	}
	if present == len(managed) {
		return false, validateExisting(root)
	}
	if present != 0 {
		return false, errors.New("incomplete existing configuration; restore missing files from your local backup or deliberately reset the entire local secret set after coordinating stored credentials; existing files were not changed")
	}
	files, err := buildFiles()
	if err != nil {
		return false, err
	}
	for _, name := range managed {
		f, err := os.OpenFile(filepath.Join(root, name), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return false, fmt.Errorf("create %s: %w; partial bootstrap must be repaired explicitly", name, err)
		}
		_, writeErr := f.Write(files[name])
		closeErr := f.Close()
		if writeErr != nil {
			return false, writeErr
		}
		if closeErr != nil {
			return false, closeErr
		}
	}
	return true, nil
}

func randomSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func digest(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:]) }

func buildFiles() (map[string][]byte, error) {
	secrets := make([]string, 8)
	for i := range secrets {
		s, err := randomSecret()
		if err != nil {
			return nil, err
		}
		secrets[i] = s
	}
	postgres, redis, bmkg, pvmbg, admin, internal := secrets[0], secrets[1], secrets[2], "pvmbg_"+secrets[3], secrets[4], secrets[5]
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	pubDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{
		"env/keys/jwt-private.pem": pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}),
		"env/keys/jwt-public.pem":  pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}),
	}
	put := func(path, content string) {
		files[path] = []byte("# Generated local development configuration. Do not commit.\n" + content)
	}
	put("env/postgres.env", "POSTGRES_DB=bnpb\nPOSTGRES_USER=bnpb\nPOSTGRES_PASSWORD="+postgres+"\n")
	put("env/redis.env", "REDIS_PASSWORD="+redis+"\n")
	put("env/bmkg-mock.env", "HTTP_ADDR=:8081\nBMKG_KEY_HASH="+digest(bmkg)+"\nGENERATION_INTERVAL=10s\nBMKG_FIXED_DELAY=100ms\n")
	put("env/pvmbg-mock.env", "HTTP_ADDR=:8082\nPVMBG_TOKEN_HASH="+digest(pvmbg)+"\nADMIN_KEY_HASH="+digest(admin)+"\nGENERATION_INTERVAL=10s\nPVMBG_DELAY_MIN=500ms\nPVMBG_DELAY_MAX=3s\n")
	put("env/auth-service.env", "HTTP_ADDR=:8090\nJWT_PRIVATE_KEY_FILE=/run/keys/jwt-private.pem\nJWT_ISSUER=bnpb-auth\nJWT_AUDIENCE=bnpb-api\nCLIENTS_FILE=/run/config/clients.json\nREDIS_ADDR=auth-store:6379\nREDIS_PASSWORD="+redis+"\nREDIS_TIMEOUT=200ms\nACCESS_TOKEN_TTL=60s\nREFRESH_TOKEN_TTL=8h\nTOKEN_RATE_LIMIT=20\nTOKEN_RATE_BURST=40\n")
	put("env/client-api.env", "HTTP_ADDR=:8080\nJWT_PUBLIC_KEY_FILE=/run/keys/jwt-public.pem\nJWT_ISSUER=bnpb-auth\nJWT_AUDIENCE=bnpb-api\nAGGREGATOR_URL=http://aggregator:9000\nINTERNAL_KEY="+internal+"\nAGGREGATOR_TIMEOUT=1500ms\nMAX_CONCURRENT=100\nRATE_LIMIT_RPS=100\nRATE_LIMIT_BURST=200\n")
	put("env/aggregator.env", "HTTP_ADDR=:9000\nDATABASE_URL=postgres://bnpb:"+postgres+"@canonical-db:5432/bnpb?sslmode=disable\nBMKG_URL=http://bmkg-mock:8081\nBMKG_API_KEY="+bmkg+"\nPVMBG_URL=http://pvmbg-mock:8082\nPVMBG_TOKEN="+pvmbg+"\nINTERNAL_KEY="+internal+"\nKAFKA_BROKERS=kafka:9092\nKAFKA_TOPIC=bnpb.hazard-events.v1\n")
	put("env/demo.env", "BMKG_API_KEY="+bmkg+"\nPVMBG_TOKEN="+pvmbg+"\nPVMBG_ADMIN_KEY="+admin+"\nINTERNAL_KEY="+internal+"\n")
	clients := []client{}
	demos := []demoClient{}
	for _, id := range []string{"media", "field-team", "bnpb-ops"} {
		secret, err := randomSecret()
		if err != nil {
			return nil, err
		}
		scopes := []string{"hazard:read:summary"}
		if id != "media" {
			scopes = append(scopes, "hazard:read:raw")
		}
		clients = append(clients, client{id, digest(secret), scopes})
		demos = append(demos, demoClient{id, secret, scopes})
	}
	files["env/clients.json"], err = json.MarshalIndent(clients, "", "  ")
	if err != nil {
		return nil, err
	}
	files["env/demo-clients.json"], err = json.MarshalIndent(demos, "", "  ")
	if err != nil {
		return nil, err
	}
	return files, nil
}

func validateExisting(root string) error {
	read := func(path string) []byte { b, _ := os.ReadFile(filepath.Join(root, path)); return b }
	privPEM, _ := pem.Decode(read("env/keys/jwt-private.pem"))
	pubPEM, _ := pem.Decode(read("env/keys/jwt-public.pem"))
	if privPEM == nil || pubPEM == nil {
		return errors.New("existing JWT keys are invalid PEM; repair explicitly")
	}
	privAny, err := x509.ParsePKCS8PrivateKey(privPEM.Bytes)
	if err != nil {
		return errors.New("existing JWT private key is invalid PKCS8")
	}
	pubAny, err := x509.ParsePKIXPublicKey(pubPEM.Bytes)
	if err != nil {
		return errors.New("existing JWT public key is invalid PKIX")
	}
	priv, ok := privAny.(ed25519.PrivateKey)
	pub, pubOK := pubAny.(ed25519.PublicKey)
	if !ok || !pubOK || !priv.Public().(ed25519.PublicKey).Equal(pub) {
		return errors.New("existing Ed25519 keys do not match")
	}
	var clients []client
	var demos []demoClient
	if json.Unmarshal(read("env/clients.json"), &clients) != nil || json.Unmarshal(read("env/demo-clients.json"), &demos) != nil || len(clients) != 3 || len(demos) != 3 {
		return errors.New("existing client JSON is incomplete or invalid")
	}
	for _, c := range clients {
		match := false
		for _, d := range demos {
			if c.ID == d.ID && c.Hash == digest(d.Secret) {
				match = true
			}
		}
		if !match || len(c.Scopes) == 0 {
			return errors.New("existing client configuration and local demo credentials disagree")
		}
	}
	required := map[string][]string{
		"postgres": {"POSTGRES_DB", "POSTGRES_USER", "POSTGRES_PASSWORD"}, "redis": {"REDIS_PASSWORD"},
		"bmkg-mock": {"HTTP_ADDR", "BMKG_KEY_HASH"}, "pvmbg-mock": {"HTTP_ADDR", "PVMBG_TOKEN_HASH", "ADMIN_KEY_HASH"},
		"auth-service": {"HTTP_ADDR", "JWT_PRIVATE_KEY_FILE", "JWT_ISSUER", "JWT_AUDIENCE", "CLIENTS_FILE", "REDIS_ADDR", "REDIS_PASSWORD"},
		"client-api":   {"HTTP_ADDR", "JWT_PUBLIC_KEY_FILE", "JWT_ISSUER", "JWT_AUDIENCE", "AGGREGATOR_URL", "INTERNAL_KEY"},
		"aggregator":   {"DATABASE_URL", "BMKG_API_KEY", "PVMBG_TOKEN", "INTERNAL_KEY", "KAFKA_BROKERS"},
		"demo":         {"BMKG_API_KEY", "PVMBG_TOKEN", "PVMBG_ADMIN_KEY", "INTERNAL_KEY"},
	}
	for name, keys := range required {
		values := map[string]string{}
		for _, line := range strings.Split(string(read("env/"+name+".env")), "\n") {
			k, v, ok := strings.Cut(line, "=")
			if ok {
				values[k] = strings.TrimSpace(v)
			}
		}
		for _, key := range keys {
			if values[key] == "" {
				return fmt.Errorf("existing env/%s.env is missing %s; repair explicitly", name, key)
			}
		}
	}
	return nil
}
