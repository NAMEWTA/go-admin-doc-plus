package version

import (
	"runtime"

	"gorm.io/gorm"

	adminmodels "go-admin/app/admin/models"
	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1786700004000MenuRouteKey)
}

var menuRouteKeys = map[string]string{
	"/admin/sys-user/index":      "system.user",
	"/admin/sys-menu/index":      "system.menu",
	"/admin/sys-role/index":      "system.role",
	"/admin/sys-dept/index":      "system.department",
	"/admin/sys-post/index":      "system.post",
	"/admin/dict/index":          "system.dictionary",
	"/admin/dict/data":           "system.dictionary-data",
	"/admin/sys-config/index":    "system.config",
	"/admin/sys-config/set":      "system.config-settings",
	"/admin/sys-api/index":       "system.api",
	"/admin/sys-login-log/index": "system.login-log",
	"/admin/sys-oper-log/index":  "system.operation-log",
	"/schedule/index":            "jobs.schedule",
	"/schedule/log":              "jobs.log",
	"/demo/product/index":        "demo.product",
	"/dev-tools/swagger/index":   "tools.swagger",
	"/dev-tools/gen/index":       "tools.generator",
	"/dev-tools/gen/editTable":   "tools.generator-edit",
	"/dev-tools/build/index":     "tools.form-builder",
	"/sys-tools/monitor":         "monitor.server",
}

func _1786700004000MenuRouteKey(db *gorm.DB, version string) error {
	if err := migrateMenuRouteKeys(db); err != nil {
		return err
	}
	return db.Create(&common.Migration{Version: version}).Error
}

func migrateMenuRouteKeys(db *gorm.DB) error {
	menu := &adminmodels.SysMenu{}
	if !db.Migrator().HasColumn(menu, "RouteKey") {
		if err := db.Migrator().AddColumn(menu, "RouteKey"); err != nil {
			return err
		}
	}
	if !db.Migrator().HasIndex(menu, "idx_sys_menu_route_key") {
		if err := db.Migrator().CreateIndex(menu, "RouteKey"); err != nil {
			return err
		}
	}
	for component, routeKey := range menuRouteKeys {
		if err := db.Table(menu.TableName()).
			Where("component = ? AND (route_key = '' OR route_key IS NULL)", component).
			Update("route_key", routeKey).Error; err != nil {
			return err
		}
	}
	return nil
}
