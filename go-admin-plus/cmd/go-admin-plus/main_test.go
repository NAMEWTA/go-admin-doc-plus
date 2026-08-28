package main

import "testing"

func TestParseOptionsKeepsServerProfilesExplicit(t *testing.T) {
	options, err := parseOptions([]string{"--profile", "server-postgres", "--listen", "127.0.0.1:9090", "--repository-root", "."})
	if err != nil {
		t.Fatal(err)
	}
	if options.profile != "server-postgres" || options.listen != "127.0.0.1:9090" {
		t.Fatalf("options = %#v", options)
	}
	if _, err := parseOptions([]string{"unexpected"}); err == nil {
		t.Fatal("parseOptions accepted positional input")
	}
}
