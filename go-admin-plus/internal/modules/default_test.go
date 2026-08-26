package modules

import (
	"reflect"
	"testing"
)

func TestDefaultModuleOrderIsExplicit(t *testing.T) {
	modules := Default().Modules()
	ids := make([]string, 0, len(modules))
	for _, module := range modules {
		ids = append(ids, module.ID())
	}
	want := []string{"runtime-queue", "admin", "demo", "jobs", "other"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("module order = %v, want %v", ids, want)
	}
}
