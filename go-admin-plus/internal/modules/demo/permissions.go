package demo

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/contracts/capabilities"
)

var moduleCapabilities = capabilities.ModuleCapabilities{
	Permissions: []capabilities.PermissionDefinition{
		{Code: PermissionProductsRead, Name: "Read demo products"},
		{Code: PermissionProductsWrite, Name: "Manage demo products"},
		{Code: PermissionProductsDelete, Name: "Delete demo products"},
	},
	Menus: []capabilities.MenuDefinition{{ID: "menu-demo-products", Key: "demo-products", Label: "Demo products", Path: "/demo/products", PermissionCode: PermissionProductsRead, SortOrder: 800}},
}

type CapabilityRegistrar interface {
	Register(context.Context, capabilities.ModuleCapabilities) error
}

func RegisterCapabilities(ctx context.Context, registrar CapabilityRegistrar) error {
	if registrar == nil {
		return ErrInternal
	}
	definitions := capabilities.ModuleCapabilities{
		Permissions: append([]capabilities.PermissionDefinition(nil), moduleCapabilities.Permissions...),
		Menus:       append([]capabilities.MenuDefinition(nil), moduleCapabilities.Menus...),
	}
	return registrar.Register(ctx, definitions)
}
