package generator

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts/capabilities"
)

var moduleCapabilities = capabilities.ModuleCapabilities{
	Permissions: []capabilities.PermissionDefinition{
		{Code: PermissionMetadataRead, Name: "Read generator metadata"},
		{Code: PermissionPreview, Name: "Preview generated modules"},
		{Code: PermissionWrite, Name: "Write generated modules"},
	},
	Menus: []capabilities.MenuDefinition{{ID: "menu-code-generator", Key: "code-generator", Label: "Code generator", Path: "/generator", PermissionCode: PermissionMetadataRead, SortOrder: 700}},
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
