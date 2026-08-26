package models

import (
	"fmt"
	"go-admin/common/global"
	seedconfig "go-admin/config"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

func InitDb(db *gorm.DB) (err error) {
	filePath := "config/db.sql"
	if global.Driver == "postgres" {
		filePath := "config/db.sql"
		if err = ExecSql(db, filePath); err != nil {
			return err
		}
		filePath = "config/pg.sql"
		err = ExecSql(db, filePath)
	} else if global.Driver == "mysql" {
		filePath = "config/db-begin-mysql.sql"
		if err = ExecSql(db, filePath); err != nil {
			return err
		}
		filePath = "config/db.sql"
		if err = ExecSql(db, filePath); err != nil {
			return err
		}
		filePath = "config/db-end-mysql.sql"
		err = ExecSql(db, filePath)
	} else {
		err = ExecSql(db, filePath)
	}
	return err
}

func ExecSql(db *gorm.DB, filePath string) error {
	sql, err := Ioutil(filePath)
	if err != nil {
		fmt.Println("数据库基础数据初始化脚本读取失败！原因:", err.Error())
		return err
	}
	sqlList := strings.Split(sql, ";")
	for i := 0; i < len(sqlList)-1; i++ {
		if strings.Contains(sqlList[i], "--") {
			fmt.Println(sqlList[i])
			continue
		}
		sql := strings.Replace(sqlList[i]+";", "\n", "", -1)
		sql = strings.TrimSpace(sql)
		if err = db.Exec(sql).Error; err != nil {
			log.Printf("error sql: %s", sql)
			if !strings.Contains(err.Error(), "Query was empty") {
				return err
			}
		}
	}
	return nil
}

func Ioutil(filePath string) (string, error) {
	clean := filepath.Clean(filePath)
	if filepath.Base(filepath.Dir(clean)) == "config" {
		contents, err := seedconfig.ReadSeedSQL(filepath.Base(clean))
		if err == nil {
			return strings.Replace(contents, "\n", "", 1), nil
		}
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.Replace(string(contents), "\n", "", 1), nil
}
