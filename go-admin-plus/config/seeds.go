package config

import (
	"embed"
	"fmt"
	"path/filepath"
)

// Seed SQL is part of the application binary so migrations do not depend on
// the process working directory or installation layout.
//
//go:embed *.sql
var seedSQL embed.FS

func ReadSeedSQL(name string) (string, error) {
	name = filepath.Base(name)
	contents, err := seedSQL.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("read embedded seed %q: %w", name, err)
	}
	return string(contents), nil
}
