package service

import (
	"testing"

	"go-admin/app/admin/models"
	"go-admin/app/admin/service/dto"
)

func TestMenuCallPreservesRouteIdentityAtEveryLevel(t *testing.T) {
	menus := []models.SysMenu{
		{MenuId: 1, ParentId: 0, MenuType: "M", RouteKey: "system.root", Component: "Layout"},
		{MenuId: 2, ParentId: 1, MenuType: "C", RouteKey: "system.user", Component: "/admin/sys-user/index"},
		{MenuId: 3, ParentId: 2, MenuType: "C", RouteKey: "system.profile", Component: "/profile/index"},
	}

	got := menuCall(&menus, menus[0])
	if got.RouteKey != "system.root" || got.Component != "Layout" {
		t.Fatalf("root identity = (%q, %q)", got.RouteKey, got.Component)
	}
	if len(got.Children) != 1 {
		t.Fatalf("root children = %d, want 1", len(got.Children))
	}
	child := got.Children[0]
	if child.RouteKey != "system.user" || child.Component != "/admin/sys-user/index" {
		t.Fatalf("child identity = (%q, %q)", child.RouteKey, child.Component)
	}
	if len(child.Children) != 1 {
		t.Fatalf("child children = %d, want 1", len(child.Children))
	}
	grandchild := child.Children[0]
	if grandchild.RouteKey != "system.profile" || grandchild.Component != "/profile/index" {
		t.Fatalf("grandchild identity = (%q, %q)", grandchild.RouteKey, grandchild.Component)
	}
}

func TestMenuUpdateCompatibilityPreservesManagedRouteKey(t *testing.T) {
	menu := models.SysMenu{
		MenuId: 2, RouteKey: "system.user", Component: "/admin/sys-user/index",
	}
	request := dto.SysMenuUpdateReq{
		MenuId: 2, Component: "/admin/sys-user/index", Title: "Updated title",
	}

	request.Generate(&menu)

	if menu.RouteKey != "system.user" {
		t.Fatalf("route key was cleared to %q", menu.RouteKey)
	}
	if menu.Component != "/admin/sys-user/index" {
		t.Fatalf("legacy component changed to %q", menu.Component)
	}
}
