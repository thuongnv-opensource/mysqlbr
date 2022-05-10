package cmd

import (
	"backup-mysql-cli/service"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

func init() {
	rootCmd.AddCommand(restore)
	restore.AddCommand(prepareBackup)
	prepareBackup.Flags().String("data-dir", "", "path to data dir")
	prepareBackup.Flags().String("backup-id", "", "id of backup snapshot")
}

var restore = &cobra.Command{
	Use:   "restore",
	Short: "restore mysql database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mysqlbr restore")
	},
}

var prepareBackup = &cobra.Command{
	Use:   "prepare",
	Short: "prepare job backup database",
	Run: func(cmd *cobra.Command, args []string) {
		dataDir, _ := cmd.Flags().GetString("data-dir")
		backupId, _ := cmd.Flags().GetString("backup-id")

		if dataDir == "" {
			log.Fatal("command need parameter --data-dir")
		}

		if backupId == "" {
			log.Fatal("command need parameter --backup-id")
		}

		service.PrepareBackup(dataDir, backupId)
	},
}
