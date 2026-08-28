package generator

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
)

var moduleCapabilities = authorization.ModuleCapabilities{
	Permissions: []authorization.PermissionDefinition{
		{Code: PermissionMetadataRead, Name: "Read generator metadata"},
		{Code: PermissionPreview, Name: "Preview generated modules"},
		{Code: PermissionWrite, Name: "Write generated modules"},
	},
	Menus: []authorization.MenuDefinition{{ID: "menu-code-generator", Key: "code-generator", Label: "Code generator", Path: "/generator", PermissionCode: PermissionMetadataRead, SortOrder: 700}},
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
