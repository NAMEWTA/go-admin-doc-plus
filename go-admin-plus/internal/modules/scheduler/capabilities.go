package scheduler

import (
	"context"

	"go-admin/internal/modules/iam/authorization"
)

var moduleCapabilities = authorization.ModuleCapabilities{
	Permissions: []authorization.PermissionDefinition{
		{Code: PermissionDefinitionsRead, Name: "Read scheduler definitions"},
		{Code: PermissionDefinitionsWrite, Name: "Manage scheduler definitions"},
		{Code: PermissionDefinitionsDelete, Name: "Delete scheduler definitions"},
		{Code: PermissionExecutionsRead, Name: "Read scheduler execution history"},
	},
	Menus: []authorization.MenuDefinition{
		{ID: "menu-scheduler-definitions", Key: "scheduler-definitions", Label: "Task schedules", Path: "/scheduler/definitions", PermissionCode: PermissionDefinitionsRead, SortOrder: 700},
		{ID: "menu-scheduler-executions", Key: "scheduler-executions", Label: "Task executions", Path: "/scheduler/executions", PermissionCode: PermissionExecutionsRead, SortOrder: 710},
	},
}

type CapabilityRegistrar interface {
	Register(context.Context, authorization.ModuleCapabilities) error
}

func RegisterCapabilities(ctx context.Context, registrar CapabilityRegistrar) error {
	if registrar == nil {
		return ErrInternal
	}
	return registrar.Register(ctx, authorization.ModuleCapabilities{
		Permissions: append([]authorization.PermissionDefinition(nil), moduleCapabilities.Permissions...),
		Menus:       append([]authorization.MenuDefinition(nil), moduleCapabilities.Menus...),
	})
}
