package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpDoesNotPrintEnvironmentSecret(t *testing.T) {
	t.Setenv("FIELD_CLI_CLIENT_SECRET", "private-test-value")
	var out, errout bytes.Buffer
	_ = run([]string{"list", "--help"}, &out, &errout)
	if strings.Contains(out.String()+errout.String(), "private-test-value") {
		t.Fatal("help disclosed environment secret")
	}
}
func TestGetFlagForms(t *testing.T) {
	for _, flag := range []string{"-api-url", "--api-url"} {
		args, err := normalizeGetArgs([]string{"hazard-id", flag, "http://localhost:8080", "--raw"})
		if err != nil || args[len(args)-1] != "hazard-id" {
			t.Fatal("value flag treated as ID", args, err)
		}
	}
}
