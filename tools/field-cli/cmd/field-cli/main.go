package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"example.com/akuanaktehat/field-cli/internal/client"
)

const usage = `field-cli reads hazards through client-api.

Commands:
  list [flags]             List hazards
  get <hazard-id> [flags]  Get one hazard

Connection flags can also be configured with FIELD_CLI_* environment variables:
  FIELD_CLI_API_URL, FIELD_CLI_AUTH_URL, FIELD_CLI_CLIENT_ID,
  FIELD_CLI_CLIENT_SECRET, FIELD_CLI_TIMEOUT
`

type connectionFlags struct {
	apiURL       string
	authURL      string
	clientID     string
	clientSecret string
	timeout      time.Duration
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "field-cli:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		_, _ = io.WriteString(stdout, usage)
		return nil
	}

	switch args[0] {
	case "list":
		return runList(args[1:], stdout, stderr)
	case "get":
		return runGet(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runList(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(stderr)
	conn, err := addConnectionFlags(fs)
	if err != nil {
		return err
	}
	typeFlag := fs.String("type", "", "SEISMIC or VOLCANIC")
	severity := fs.String("severity", "", "NORMAL, WASPADA, SIAGA, or AWAS")
	since := fs.String("since", "", "occurred_at lower bound in RFC3339")
	limit := fs.Int("limit", 100, "number of hazards to request (1-500)")
	cursor := fs.String("cursor", "", "opaque pagination cursor")
	raw := fs.Bool("raw", false, "request raw fields; requires hazard:read:raw")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("list does not accept positional arguments")
	}
	api, err := newAPIClient(conn)
	if err != nil {
		return err
	}
	body, err := api.List(context.Background(), client.ListOptions{
		Type:     *typeFlag,
		Severity: *severity,
		Since:    *since,
		Limit:    *limit,
		Cursor:   *cursor,
		Raw:      *raw,
	})
	if err != nil {
		return err
	}
	return client.PrintJSON(stdout, body)
}

func runGet(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("get", flag.ContinueOnError)
	fs.SetOutput(stderr)
	conn, err := addConnectionFlags(fs)
	if err != nil {
		return err
	}
	raw := fs.Bool("raw", false, "request raw fields; requires hazard:read:raw")
	normalized, err := normalizeGetArgs(args)
	if err != nil {
		return err
	}
	if err := fs.Parse(normalized); err != nil {
		return err
	}
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: field-cli get <hazard-id> [flags]")
	}
	api, err := newAPIClient(conn)
	if err != nil {
		return err
	}
	body, err := api.Get(context.Background(), fs.Arg(0), *raw)
	if err != nil {
		return err
	}
	return client.PrintJSON(stdout, body)
}

func normalizeGetArgs(args []string) ([]string, error) {
	valueFlags := map[string]bool{
		"--api-url":       true,
		"--auth-url":      true,
		"--client-id":     true,
		"--client-secret": true,
		"--timeout":       true,
	}
	flags := make([]string, 0, len(args))
	id := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			name := strings.SplitN(arg, "=", 2)[0]
			if valueFlags[name] && !strings.Contains(arg, "=") {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag %s requires a value", name)
				}
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		if id != "" {
			return nil, errors.New("get accepts exactly one hazard-id")
		}
		id = arg
	}
	if id != "" {
		flags = append(flags, id)
	}
	return flags, nil
}

func addConnectionFlags(fs *flag.FlagSet) (*connectionFlags, error) {
	timeout, err := durationEnv("FIELD_CLI_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	flags := &connectionFlags{
		apiURL:       envOr("FIELD_CLI_API_URL", "http://127.0.0.1:8080"),
		authURL:      envOr("FIELD_CLI_AUTH_URL", "http://127.0.0.1:8090"),
		clientID:     envOr("FIELD_CLI_CLIENT_ID", "field-team"),
		clientSecret: os.Getenv("FIELD_CLI_CLIENT_SECRET"),
		timeout:      timeout,
	}
	fs.StringVar(&flags.apiURL, "api-url", flags.apiURL, "client-api base URL")
	fs.StringVar(&flags.authURL, "auth-url", flags.authURL, "auth-service base URL")
	fs.StringVar(&flags.clientID, "client-id", flags.clientID, "OAuth client ID")
	fs.StringVar(&flags.clientSecret, "client-secret", flags.clientSecret, "OAuth client secret")
	fs.DurationVar(&flags.timeout, "timeout", flags.timeout, "per-request timeout")
	return flags, nil
}

func newAPIClient(flags *connectionFlags) (*client.Client, error) {
	if flags.clientID == "" || flags.clientSecret == "" {
		return nil, errors.New("client credentials are required via FIELD_CLI_CLIENT_SECRET or --client-secret")
	}
	if flags.timeout <= 0 {
		return nil, errors.New("timeout must be positive")
	}
	return client.New(client.Config{
		APIURL:       flags.apiURL,
		AuthURL:      flags.authURL,
		ClientID:     flags.clientID,
		ClientSecret: flags.clientSecret,
		Timeout:      flags.timeout,
	})
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration: %w", name, err)
	}
	return parsed, nil
}
