package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "docker-compose.yml"), []byte("services: {}"), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestGenerateIsIdempotentAndValid(t *testing.T) {
	root := fixture(t)
	created, err := generate(root)
	if err != nil || !created {
		t.Fatalf("generate: %v, %v", created, err)
	}
	before := map[string][]byte{}
	for _, name := range managed {
		b, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		before[name] = b
	}
	created, err = generate(root)
	if err != nil || created {
		t.Fatalf("second generate: %v, %v", created, err)
	}
	for _, name := range managed {
		b, _ := os.ReadFile(filepath.Join(root, name))
		if !bytes.Equal(before[name], b) {
			t.Fatalf("rotated %s", name)
		}
	}
}

func TestPartialConfigurationDoesNotRotate(t *testing.T) {
	root := fixture(t)
	if _, err := generate(root); err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(root, "env/keys/jwt-private.pem")
	before, _ := os.ReadFile(keyPath)
	if err := os.Remove(filepath.Join(root, "env/redis.env")); err != nil {
		t.Fatal(err)
	}
	if _, err := generate(root); err == nil {
		t.Fatal("accepted incomplete configuration")
	}
	after, _ := os.ReadFile(keyPath)
	if !bytes.Equal(before, after) {
		t.Fatal("rotated key after partial failure")
	}
}

func TestRejectsMismatchedExistingKey(t *testing.T) {
	root := fixture(t)
	if _, err := generate(root); err != nil {
		t.Fatal(err)
	}
	other, err := buildFiles()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "env/keys/jwt-public.pem"), other["env/keys/jwt-public.pem"], 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := generate(root); err == nil {
		t.Fatal("accepted mismatched keys")
	}
}
