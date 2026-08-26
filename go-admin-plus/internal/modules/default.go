package modules

import (
	"go-admin/app/admin"
	"go-admin/app/demo"
	"go-admin/app/other"
	"go-admin/internal/application"
)

func Default() application.ModuleSet {
	return application.NewModuleSet(
		newRuntimeQueueModule(),
		admin.NewModule(),
		demo.NewModule(),
		newJobsModule(),
		other.NewModule(),
	)
}
