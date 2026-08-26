package version

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type menuBeforeRouteKey struct {
	MenuId    int    `gorm:"column:menu_id;primaryKey"`
	Component string `gorm:"column:component"`
}

func (menuBeforeRouteKey) TableName() string { return "sys_menu" }

func TestMenuRouteKeyMigrationExpandsWithoutRemovingComponent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := db.AutoMigrate(&menuBeforeRouteKey{}); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	rows := []menuBeforeRouteKey{
		{MenuId: 1, Component: "/admin/sys-user/index"},
		{MenuId: 2, Component: "/schedule/index"},
		{MenuId: 3, Component: "/custom/page"},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed menus: %v", err)
	}

	if err := migrateMenuRouteKeys(db); err != nil {
		t.Fatalf("migrate route keys: %v", err)
	}

	var got []struct {
		MenuId    int
		Component string
		RouteKey  string
	}
	if err := db.Table("sys_menu").Order("menu_id").Scan(&got).Error; err != nil {
		t.Fatalf("read migrated menus: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d menus, want 3", len(got))
	}
	if got[0].RouteKey != "system.user" || got[1].RouteKey != "jobs.schedule" || got[2].RouteKey != "" {
		t.Fatalf("route keys = %#v", got)
	}
	for index := range rows {
		if got[index].Component != rows[index].Component {
			t.Errorf("menu %d component changed from %q to %q", rows[index].MenuId, rows[index].Component, got[index].Component)
		}
	}
}

func TestMenuRouteKeyMigrationIsRepeatableAndPreservesExplicitKeys(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open SQLite: %v", err)
	}
	if err := db.AutoMigrate(&menuBeforeRouteKey{}); err != nil {
		t.Fatalf("create old schema: %v", err)
	}
	if err := db.Create(&menuBeforeRouteKey{MenuId: 1, Component: "/demo/product/index"}).Error; err != nil {
		t.Fatalf("seed menu: %v", err)
	}
	if err := migrateMenuRouteKeys(db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := db.Table("sys_menu").Where("menu_id = ?", 1).Update("route_key", "demo.custom").Error; err != nil {
		t.Fatalf("set explicit key: %v", err)
	}
	if err := migrateMenuRouteKeys(db); err != nil {
		t.Fatalf("second migration: %v", err)
	}

	var routeKey string
	if err := db.Table("sys_menu").Select("route_key").Where("menu_id = ?", 1).Scan(&routeKey).Error; err != nil {
		t.Fatalf("read route key: %v", err)
	}
	if routeKey != "demo.custom" {
		t.Fatalf("explicit route key was overwritten with %q", routeKey)
	}
}
