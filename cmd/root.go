package cmd

import (
	"fmt"

	"github.com/cvetkovski98/zvax-auth/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string
var root = &cobra.Command{
	Short: "Auth microservice",
	Long:  `Auth microservice`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Auth microservice")
	},
}

func init() {
	cobra.OnInitialize(configure)
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.dev.yml", "config file name")
	root.AddCommand(runCommand)
	root.AddCommand(migrateCommand)
}

func configure() {
	if err := config.LoadConfig(cfgFile); err != nil {
		panic(err)
	}
}

func Execute() error {
	return root.Execute()
}
