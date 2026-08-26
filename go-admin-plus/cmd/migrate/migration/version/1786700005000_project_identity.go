package version

import (
	"runtime"

	"gorm.io/gorm"

	"go-admin/cmd/migrate/migration"
	common "go-admin/common/models"
)

func init() {
	_, fileName, _, _ := runtime.Caller(0)
	migration.Migrate.SetVersion(migration.GetFilename(fileName), _1786700005000ProjectIdentity)
}

func _1786700005000ProjectIdentity(db *gorm.DB, version string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migrateProjectIdentity(tx); err != nil {
			return err
		}
		return tx.Create(&common.Migration{Version: version}).Error
	})
}

func migrateProjectIdentity(db *gorm.DB) error {
	if err := db.Table("sys_config").
		Where("config_key = ? AND config_value IN ?", "sys_app_name", []string{
			"go-admin",
			"go-admin管理系统",
		}).
		Update("config_value", "Go Admin Plus").Error; err != nil {
		return err
	}

	if err := db.Table("sys_config").
		Where("config_key = ? AND config_value IN ?", "sys_app_logo", []string{
			"https://doc-image.zhangwj.com/img/go-admin.png",
			"https://gitee.com/mydearzwj/image/raw/master/img/go-admin.png",
		}).
		Update("config_value", "").Error; err != nil {
		return err
	}

	return db.Table("sys_user").
		Where("username = ? AND nick_name = ? AND phone = ? AND email = ?", "admin", "zhangwj", "13818888888", "1@qq.com").
		Updates(map[string]interface{}{
			"nick_name": "Administrator",
			"phone":     "",
			"email":     "",
		}).Error
}
