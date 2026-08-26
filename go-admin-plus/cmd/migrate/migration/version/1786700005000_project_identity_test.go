package version

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type identityConfig struct {
	ID          int `gorm:"primaryKey"`
	ConfigKey   string
	ConfigValue string
}

func (identityConfig) TableName() string { return "sys_config" }

type identityUser struct {
	ID       int `gorm:"column:user_id;primaryKey"`
	Username string
	NickName string
	Phone    string
	Email    string
}

func (identityUser) TableName() string { return "sys_user" }

func TestProjectIdentityMigrationUpdatesOnlyLegacyDefaults(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&identityConfig{}, &identityUser{}); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	rows := []identityConfig{
		{ID: 1, ConfigKey: "sys_app_name", ConfigValue: "go-admin管理系统"},
		{ID: 2, ConfigKey: "sys_app_logo", ConfigValue: "https://doc-image.zhangwj.com/img/go-admin.png"},
		{ID: 3, ConfigKey: "sys_app_name", ConfigValue: "Customer Console"},
		{ID: 4, ConfigKey: "unrelated", ConfigValue: "go-admin"},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	users := []identityUser{
		{ID: 1, Username: "admin", NickName: "zhangwj", Phone: "13818888888", Email: "1@qq.com"},
		{ID: 2, Username: "custom", NickName: "zhangwj", Phone: "13818888888", Email: "1@qq.com"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed users: %v", err)
	}

	for i := 0; i < 2; i++ {
		if err := migrateProjectIdentity(db); err != nil {
			t.Fatalf("migrate attempt %d: %v", i+1, err)
		}
	}

	var got []identityConfig
	if err := db.Order("id").Find(&got).Error; err != nil {
		t.Fatalf("read result: %v", err)
	}
	want := []string{"Go Admin Plus", "", "Customer Console", "go-admin"}
	for i := range want {
		if got[i].ConfigValue != want[i] {
			t.Errorf("row %d value = %q, want %q", got[i].ID, got[i].ConfigValue, want[i])
		}
	}

	var gotUsers []identityUser
	if err := db.Order("user_id").Find(&gotUsers).Error; err != nil {
		t.Fatalf("read users: %v", err)
	}
	if gotUsers[0].NickName != "Administrator" || gotUsers[0].Phone != "" || gotUsers[0].Email != "" {
		t.Errorf("legacy administrator identity was not replaced: %+v", gotUsers[0])
	}
	if gotUsers[1] != users[1] {
		t.Errorf("non-administrator identity changed: got %+v want %+v", gotUsers[1], users[1])
	}
}
