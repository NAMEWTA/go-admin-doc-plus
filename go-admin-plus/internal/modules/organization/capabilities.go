package organization

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts/capabilities"
)

var moduleCapabilities = capabilities.ModuleCapabilities{
	Permissions: []capabilities.PermissionDefinition{
		{Code: PermissionDepartmentsRead, Name: "Read organization departments"},
		{Code: PermissionDepartmentsWrite, Name: "Manage organization departments"},
		{Code: PermissionDepartmentsDelete, Name: "Delete organization departments"},
		{Code: PermissionPositionsRead, Name: "Read organization positions"},
		{Code: PermissionPositionsWrite, Name: "Manage organization positions"},
		{Code: PermissionPositionsDelete, Name: "Delete organization positions"},
	},
	Menus: []capabilities.MenuDefinition{
		{ID: "menu-organization-departments", Key: "organization-departments", Label: "Departments", Path: "/organization/departments", PermissionCode: PermissionDepartmentsRead, SortOrder: 600},
		{ID: "menu-organization-positions", Key: "organization-positions", Label: "Positions", Path: "/organization/positions", PermissionCode: PermissionPositionsRead, SortOrder: 610},
	},
}

type CapabilityRegistrar interface {
	Register(context.Context, capabilities.ModuleCapabilities) error
}

func RegisterCapabilities(ctx context.Context, registrar CapabilityRegistrar) error {
	if registrar == nil {
		return ErrInternal
	}
	return registrar.Register(ctx, capabilities.ModuleCapabilities{
		Permissions: append([]capabilities.PermissionDefinition(nil), moduleCapabilities.Permissions...),
		Menus:       append([]capabilities.MenuDefinition(nil), moduleCapabilities.Menus...),
	})
}
