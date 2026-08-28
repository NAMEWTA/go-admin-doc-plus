package settings

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts/capabilities"
)

var moduleCapabilities = capabilities.ModuleCapabilities{
	Permissions: []capabilities.PermissionDefinition{
		{Code: PermissionValuesRead, Name: "Read settings"}, {Code: PermissionValuesWrite, Name: "Manage settings"}, {Code: PermissionValuesDelete, Name: "Delete settings"},
		{Code: PermissionDictionariesRead, Name: "Read dictionaries"}, {Code: PermissionDictionariesWrite, Name: "Manage dictionaries"}, {Code: PermissionDictionariesDelete, Name: "Delete dictionaries"},
		{Code: PermissionOptionsRead, Name: "Read dictionary options"},
	},
	Menus: []capabilities.MenuDefinition{
		{ID: "menu-settings-values", Key: "settings-values", Label: "Settings", Path: "/settings/values", PermissionCode: PermissionValuesRead, SortOrder: 600},
		{ID: "menu-settings-dictionaries", Key: "settings-dictionaries", Label: "Dictionaries", Path: "/settings/dictionaries", PermissionCode: PermissionDictionariesRead, SortOrder: 610},
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
