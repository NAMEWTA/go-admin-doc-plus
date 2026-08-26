package characterization_test

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPrePhaseOneSQLiteFixture(t *testing.T) {
	const wantSHA256 = "07810c8bed64f4dd1921d541402a695d76109049d6f8117f9d49deb19c61a81c"

	fixture, err := os.ReadFile("testdata/pre_phase1.sql")
	if err != nil {
		t.Fatalf("read pre-phase-one fixture: %v", err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(fixture)); got != wantSHA256 {
		t.Fatalf("fixture SHA-256 = %s, want %s", got, wantSHA256)
	}

	contract := string(fixture)
	for _, required := range []string{
		"CREATE TABLE sys_migration",
		"CREATE TABLE demo_product",
		"CREATE INDEX idx_demo_product_create_by",
		"1786700000000",
		"PRE-PHASE1-001",
		"deleted_at BIGINT NOT NULL DEFAULT 0",
	} {
		if !strings.Contains(contract, required) {
			t.Errorf("fixture omitted %q", required)
		}
	}
}
