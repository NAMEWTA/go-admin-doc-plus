package settings

import (
	"context"

	"go-admin/internal/modules/iam/authorization"
)

var moduleCapabilities = authorization.ModuleCapabilities{
	Permissions: []authorization.PermissionDefinition{
		{Code: PermissionValuesRead, Name: "Read settings"}, {Code: PermissionValuesWrite, Name: "Manage settings"}, {Code: PermissionValuesDelete, Name: "Delete settings"},
		{Code: PermissionDictionariesRead, Name: "Read dictionaries"}, {Code: PermissionDictionariesWrite, Name: "Manage dictionaries"}, {Code: PermissionDictionariesDelete, Name: "Delete dictionaries"},
		{Code: PermissionOptionsRead, Name: "Read dictionary options"},
	},
	Menus: []authorization.MenuDefinition{
		{ID: "menu-settings-values", Key: "settings-values", Label: "Settings", Path: "/settings/values", PermissionCode: PermissionValuesRead, SortOrder: 600},
		{ID: "menu-settings-dictionaries", Key: "settings-dictionaries", Label: "Dictionaries", Path: "/settings/dictionaries", PermissionCode: PermissionDictionariesRead, SortOrder: 610},
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
