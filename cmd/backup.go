package cmd

import (
	"backup-mysql-cli/service"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

func init() {
	scheduleBackup.Flags().String("config", "", "Add schedule config file")
	lsBackup.Flags().String("data-dir", "", "path to data dir")
	createBackup.Flags().String("config", "", "path to config file")
	rootCmd.AddCommand(backup)
	backup.AddCommand(scheduleBackup)
	backup.AddCommand(lsBackup)
	backup.AddCommand(createBackup)
}

var backup = &cobra.Command{
	Use:   "backup",
	Short: "backup mysql database",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("mysqlbr schedule - Lên lịch backup database tự động\nmysqlbr list - Liệt kê danh sách các bản backup")
	},
}

var scheduleBackup = &cobra.Command{
	Use:   "schedule",
	Short: "schedule job backup database",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile == "" {
			log.Fatal("command need parameter --config")
		}
		config := service.ParseConfigFile(configFile)
		service.ScheduleBackup(config)
	},
}

var lsBackup = &cobra.Command{
	Use:   "ls",
	Short: "ls job backup database",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("data-dir")
		lsBackupExec(configFile)
	},
}

var createBackup = &cobra.Command{
	Use:   "create",
	Short: "create job backup database",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		if configFile == "" {
			log.Fatal("command need parameter --config")
		}
		config := service.ParseConfigFile(configFile)

		service.MakeOneSnapshotBackup(config)
	},
}

func lsBackupExec(dataDir string) {
	if dataDir == "" {
		log.Fatal("command need parameter --data-dir")
	}
	var resultFilePath string

	if strings.HasSuffix(dataDir, "/") {
		resultFilePath = dataDir + "bk.dat"
	} else {
		resultFilePath = dataDir + "/bk.dat"
	}

	if strings.Contains(resultFilePath, "@") {
		// ssh
		err, f := service.RsyncCopyFile(resultFilePath, "", "")
		if err != nil {
			log.Fatal(err)
		}
		resultFilePath = f
	}

	service.ListBackup(resultFilePath, dataDir)
}
