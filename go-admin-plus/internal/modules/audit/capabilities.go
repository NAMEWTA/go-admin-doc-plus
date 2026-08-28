package audit

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
)

var moduleCapabilities = authorization.ModuleCapabilities{
	Permissions: []authorization.PermissionDefinition{
		{Code: string(PermissionRead), Name: "Read audit records"},
		{Code: string(PermissionCleanup), Name: "Clean up audit records"},
	},
	Menus: []authorization.MenuDefinition{{
		ID: "menu-audit-records", Key: "audit-records", Label: "Audit records", Path: "/audit/records",
		PermissionCode: string(PermissionRead), SortOrder: 500,
	}},
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
