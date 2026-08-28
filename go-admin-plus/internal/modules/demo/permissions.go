package demo

import (
	"context"

	"github.com/NAMEWTA/go-admin-plus/go-admin-plus/internal/modules/iam/authorization"
)

var moduleCapabilities = authorization.ModuleCapabilities{
	Permissions: []authorization.PermissionDefinition{
		{Code: PermissionProductsRead, Name: "Read demo products"},
		{Code: PermissionProductsWrite, Name: "Manage demo products"},
		{Code: PermissionProductsDelete, Name: "Delete demo products"},
	},
	Menus: []authorization.MenuDefinition{{ID: "menu-demo-products", Key: "demo-products", Label: "Demo products", Path: "/demo/products", PermissionCode: PermissionProductsRead, SortOrder: 800}},
}

type CapabilityRegistrar interface {
	Register(context.Context, authorization.ModuleCapabilities) error
}

func RegisterCapabilities(ctx context.Context, registrar CapabilityRegistrar) error {
	if registrar == nil {
		return ErrInternal
	}
	capabilities := authorization.ModuleCapabilities{
		Permissions: append([]authorization.PermissionDefinition(nil), moduleCapabilities.Permissions...),
		Menus:       append([]authorization.MenuDefinition(nil), moduleCapabilities.Menus...),
	}
	return registrar.Register(ctx, capabilities)
}
