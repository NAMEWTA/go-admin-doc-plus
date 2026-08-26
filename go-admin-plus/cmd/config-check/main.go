// Command config-check validates a server runtime profile without opening a
// listener or connecting to dependencies.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	runtimeconfig "go-admin/internal/platform/config"
)

func main() {
	os.Exit(run(os.Args[1:], environmentMap(os.Environ()), os.Stdout, os.Stderr))
}

func run(args []string, environment map[string]string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("config-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profileName := flags.String("profile", "", "runtime profile")
	file := flags.String("config", "", "JSON runtime configuration file")
	var httpListen, logLevel, sqlitePath optionalString
	flags.Var(&httpListen, "http-listen", "non-sensitive HTTP listen override")
	flags.Var(&logLevel, "log-level", "non-sensitive log level override")
	flags.Var(&sqlitePath, "sqlite-path", "non-sensitive Server SQLite path override")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "config-check: positional arguments are not permitted")
		return 2
	}
	profile := runtimeconfig.Profile(*profileName)
	cli := make(map[string]string, 3)
	if httpListen.set {
		cli["http.listen"] = httpListen.value
	}
	if logLevel.set {
		cli["log.level"] = logLevel.value
	}
	if sqlitePath.set {
		cli["database.path"] = sqlitePath.value
	}
	snapshot, err := runtimeconfig.Load(runtimeconfig.Input{
		Profile:     profile,
		File:        *file,
		Environment: environment,
		CLI:         cli,
	})
	if err != nil {
		fmt.Fprintf(stderr, "config-check: %v\n", err)
		return 1
	}
	if err := json.NewEncoder(stdout).Encode(map[string]string{
		"profile": string(snapshot.Profile()),
		"status":  "valid",
	}); err != nil {
		fmt.Fprintln(stderr, "config-check: write result failed")
		return 1
	}
	return 0
}

type optionalString struct {
	value string
	set   bool
}

func (value *optionalString) String() string { return value.value }

func (value *optionalString) Set(candidate string) error {
	value.value = candidate
	value.set = true
	return nil
}

func environmentMap(entries []string) map[string]string {
	environment := make(map[string]string)
	for _, entry := range entries {
		key, value, found := strings.Cut(entry, "=")
		if found {
			environment[key] = value
		}
	}
	return environment
}
