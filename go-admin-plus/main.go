package main

import (
	"go-admin/cmd"
)

//go:generate swag init --parseDependency --parseDepth=6 --instanceName admin -o ./docs/admin

// @title Go Admin Plus API
// @version 2.0.0
// @description Go Admin Plus 管理平台的 HTTP API。
// @license.name MIT
// @license.url https://github.com/NAMEWTA/go-admin-plus/blob/main/LICENSE

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	cmd.Execute()
}
