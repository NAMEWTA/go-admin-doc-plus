package migrate

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"text/template"
	"time"

	"github.com/go-admin-team/go-admin-core/v2/config/source/file"
	"github.com/go-admin-team/go-admin-core/v2/sdk"
	"github.com/go-admin-team/go-admin-core/v2/sdk/pkg"
	"github.com/spf13/cobra"
	"gorm.io/gorm"

	"github.com/go-admin-team/go-admin-core/v2/sdk/config"
	"go-admin/cmd/migrate/migration"
	_ "go-admin/cmd/migrate/migration/version"
	_ "go-admin/cmd/migrate/migration/version-local"
	"go-admin/common/database"
)

var (
	configYml string
	generate  bool
	goAdmin   bool
	host      string
	StartCmd  = &cobra.Command{
		Use:     "migrate",
		Short:   "Initialize the database",
		Example: "go-admin migrate -c config/settings.yml",
		RunE:    func(cmd *cobra.Command, args []string) error { return run() },
	}
)

// fixme 在您看不见代码的时候运行迁移，我觉得是不安全的，所以编译后最好不要去执行迁移
func init() {
	StartCmd.PersistentFlags().StringVarP(&configYml, "config", "c", "config/settings.yml", "Start server with provided configuration file")
	StartCmd.PersistentFlags().BoolVarP(&generate, "generate", "g", false, "generate migration file")
	StartCmd.PersistentFlags().BoolVarP(&goAdmin, "goAdmin", "a", false, "generate go-admin migration file")
	StartCmd.PersistentFlags().StringVarP(&host, "domain", "d", "*", "select tenant host")
}

func run() error {
	if !generate {
		fmt.Println(`start init`)
		var migrationErr error
		config.Setup(
			file.NewSource(file.WithPath(configYml)),
			func() { migrationErr = initDB() },
		)
		return migrationErr
	}
	fmt.Println(`generate migration file`)
	return genFile()
}

func migrateModel() error {
	if host == "" {
		host = "*"
	}
	db := sdk.Runtime.GetDbByTenant(host)
	if db == nil {
		if len(sdk.Runtime.GetAllDb()) == 1 && host == "*" {
			for k, v := range sdk.Runtime.GetAllDb() {
				db = v
				host = k
				break
			}
		}
	}
	if db == nil {
		return fmt.Errorf("未找到数据库配置")
	}
	driver := ""
	if databaseConfig := config.DatabasesConfig[host]; databaseConfig != nil {
		driver = databaseConfig.Driver
	}
	return migrateDatabase(context.Background(), db, driver, migration.Migrate)
}

func migrateDatabase(ctx context.Context, db *gorm.DB, driver string, runner *migration.Migration) error {
	if driver == "mysql" {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4")
	}
	_, err := runner.Run(ctx, db)
	return err
}

func initDB() error {
	database.Setup()
	fmt.Println("数据库迁移开始")
	if err := migrateModel(); err != nil {
		return err
	}
	fmt.Println(`数据库基础数据初始化成功`)
	return nil
}

func genFile() error {
	t1, err := template.ParseFiles("template/migrate.template")
	if err != nil {
		return err
	}
	m := map[string]string{}
	m["GenerateTime"] = strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	m["Package"] = "version_local"
	if goAdmin {
		m["Package"] = "version"
	}
	var b1 bytes.Buffer
	err = t1.Execute(&b1, m)
	if goAdmin {
		pkg.FileCreate(b1, "./cmd/migrate/migration/version/"+m["GenerateTime"]+"_migrate.go")
	} else {
		pkg.FileCreate(b1, "./cmd/migrate/migration/version-local/"+m["GenerateTime"]+"_migrate.go")
	}
	return nil
}
