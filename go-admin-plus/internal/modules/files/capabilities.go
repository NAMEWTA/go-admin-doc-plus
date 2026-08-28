package files

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
)

var moduleCapabilities = authorization.ModuleCapabilities{
	Permissions: []authorization.PermissionDefinition{
		{Code: PermissionFilesRead, Name: "Read files"},
		{Code: PermissionFilesWrite, Name: "Upload files"},
		{Code: PermissionFilesDelete, Name: "Delete files"},
	},
	Menus: []authorization.MenuDefinition{{ID: "menu-files-objects", Key: "files-objects", Label: "Files", Path: "/files", PermissionCode: PermissionFilesRead, SortOrder: 900}},
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
