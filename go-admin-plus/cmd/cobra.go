package cmd

import (
	"errors"
	"fmt"
	"github.com/go-admin-team/go-admin-core/v2/sdk/pkg"
	"go-admin/cmd/app"
	"go-admin/common/global"
	"os"

	"github.com/spf13/cobra"

	"go-admin/cmd/api"
	"go-admin/cmd/config"
	"go-admin/cmd/migrate"
	"go-admin/cmd/version"
)

var rootCmd = &cobra.Command{
	Use:          "go-admin",
	Short:        "Go Admin Plus command line",
	SilenceUsage: true,
	Long:         `Go Admin Plus command line`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			tip()
			return errors.New(pkg.Red("requires at least one arg"))
		}
		return nil
	},
	PersistentPreRunE: func(*cobra.Command, []string) error { return nil },
	Run: func(cmd *cobra.Command, args []string) {
		tip()
	},
}

func tip() {
	usageStr := `欢迎使用 ` + pkg.Green(`Go Admin Plus `+global.Version) + `，使用 ` + pkg.Red(`-h`) + ` 查看命令`
	usageStr1 := `项目文档：https://github.com/NAMEWTA/go-admin-plus`
	fmt.Printf("%s\n", usageStr)
	fmt.Printf("%s\n", usageStr1)
}

func init() {
	rootCmd.AddCommand(api.StartCmd)
	rootCmd.AddCommand(migrate.StartCmd)
	rootCmd.AddCommand(version.StartCmd)
	rootCmd.AddCommand(config.StartCmd)
	rootCmd.AddCommand(app.StartCmd)
}

// Execute : apply commands
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}
