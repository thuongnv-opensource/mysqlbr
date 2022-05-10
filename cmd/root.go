package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "mysqlbr",
	Short: "mysqlbr use to manage mysql backup/restore",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		fmt.Println("mysqlbr -h to show command")
	},
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
