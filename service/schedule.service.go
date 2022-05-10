package service

import (
	"backup-mysql-cli/database"
	"backup-mysql-cli/utils"
	"fmt"
	"github.com/robfig/cron"
	"gorm.io/gorm"
	"io/fs"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func ScheduleBackup(config ScheduleConfig) {
	var quit = make(chan struct{})

	var channel = make(chan int, 1)
	channel <- 1

	// create backup report file
	utils.CreateFileIfNotExist(config.Backup.DataDir + "/bk.dat")
	db := database.GetInstance(config.Mysql.Host, config.Mysql.Port, config.Mysql.User, config.Mysql.Password)
	initDatabase(db)

	for _, strategy := range config.Backup.Strategies {
		utils.PrintBlue("[SETUP BACKUP SCHEDULE]")
		fmt.Printf(": %s cron: %s\n", strategy.Name, strategy.Cron)
		time.Sleep(10 * time.Second)

		go func(ss ScheduleStrategy) {
			c := cron.New()
			c.AddFunc(ss.Cron, func() {
				execScheduleBackup(ss, config.Backup.DataDir, channel, config.Mysql, db)
			})
			c.Start()
		}(strategy)

		time.Sleep(2 * time.Second)
	}

	// prevent app exit
	<-quit
}

func ExecRawQuery(db *gorm.DB, query string) bool {
	tx := db.Exec(query)

	if tx.Error != nil {
		return false
	}

	return true
}

func initDatabase(db *gorm.DB) {
	if ExecRawQuery(db, "CREATE SCHEMA `mysqlbr_metadata`;") {
		fmt.Println("[MYSQL]: CREATE SCHEMA mysqlbr_metadata")
		ExecRawQuery(db, "CREATE TABLE `mysqlbr_metadata`.`last_backup_record` (`backup_id` VARCHAR(50) NOT NULL, PRIMARY KEY (`backup_id`));")
		fmt.Println("[MYSQL]: CREATE TABLE `mysqlbr_metadata`")
		ExecRawQuery(db, "INSERT INTO `mysqlbr_metadata`.`last_backup_record` (`backup_id`) VALUES ('init');")
		fmt.Println("[MYSQL]: INSERT INTO `mysqlbr_metadata`")
	}
}

func updateLastBackupRecordId(db *gorm.DB, backupId string) {
	if ExecRawQuery(db, fmt.Sprintf("UPDATE `mysqlbr_metadata`.`last_backup_record` SET backup_id='%s';", backupId)) {
		utils.PrintBlue("[MYSQL]")
		fmt.Println(" update last backup id to: " + backupId)
	} else {
		utils.PrintRed("[ERROR]")
		fmt.Println(" update last backup id to: " + backupId)
	}
}

func execScheduleBackup(strategy ScheduleStrategy, dataDir string, channel chan int, mysqlInfo MysqlConnectionInfo, db *gorm.DB) {
	utils.PrintPurple("[START BACKUP BASE]")
	fmt.Printf(": \"%s\" at: \"%s\"\n", strategy.Name, strings.Split(time.Now().String(), ".")[0])

	<-channel

	timestamp := utils.MakeTimestampFromNow()
	backupRecordId := strconv.FormatInt(timestamp, 10)
	baseStorePath := fmt.Sprintf("%s/%s/%s/base", dataDir, strategy.Name, backupRecordId)
	utils.MakeDirIfNotExist(baseStorePath)

	cmdFlags := fmt.Sprintf("--backup --target-dir=%s", baseStorePath)
	cmdFlags = cmdFlags + ParseMysqlConnectInfoToXtrabackupFlags(mysqlInfo)
	if strategy.ExecFlags != "" {
		cmdFlags = cmdFlags + " " + strategy.ExecFlags
	}
	updateLastBackupRecordId(db, backupRecordId)

	c := exec.Command("xtrabackup", strings.Split(cmdFlags, " ")...)
	err, o := utils.ExecCommand(c)

	utils.CreateFileIfNotExist(baseStorePath + "/result.log")
	ioutil.WriteFile(baseStorePath+"/result.log", []byte(c.String()+"\n"+o), fs.ModeAppend)

	utils.AddBackupReportToFile(dataDir+"/bk.dat", baseStorePath, strategy.Name, timestamp, 0, err)

	channel <- 1

	if err != nil {
		return
	}

	// handle incrementalBackup
	incrementalCron := cron.New()
	for _, incrementalBackup := range strategy.IncrementalBackups {
		incrementalCron.AddFunc("@every "+incrementalBackup.Every, func() {

			<-channel

			// find to total incremental with base timestamp
			count := 0
			for _, record := range utils.ParseBackupReportFile(dataDir + "/bk.dat") {
				if record.Incremental != nil && utils.MakeTimestamp(*record.Base) == timestamp {
					count++
				}
			}

			stopIncremental := false

			if count == incrementalBackup.Take-1 {
				incrementalCron.Stop()
				stopIncremental = true
			}

			if count > incrementalBackup.Take {
				utils.PrintBlue("[STOP INCREMENTAL]")
				fmt.Println(" with base:" + strconv.FormatInt(timestamp, 10))
				incrementalCron.Stop()
				return
			}

			utils.PrintPurple("[START INCREMENTAL BACKUP]")
			fmt.Printf(": \"%s\" at: \"%s\"\n", strconv.FormatInt(timestamp, 10), strings.Split(time.Now().String(), ".")[0])
			incrementalTimestamp := utils.MakeTimestampFromNow()
			backupRecordId := strconv.FormatInt(incrementalTimestamp, 10)

			incrementalStorePath := fmt.Sprintf("%s/%s/%s/incremental/%s", dataDir, strategy.Name, strconv.FormatInt(timestamp, 10), backupRecordId)
			utils.MakeDirIfNotExist(incrementalStorePath)
			cmdFlags := fmt.Sprintf("--backup --target-dir=%s --incremental-basedir=%s", incrementalStorePath, baseStorePath)
			cmdFlags = cmdFlags + ParseMysqlConnectInfoToXtrabackupFlags(mysqlInfo)
			if strategy.ExecFlags != "" {
				cmdFlags = cmdFlags + " " + strings.TrimSpace(strategy.ExecFlags)
			}
			baseStorePath = incrementalStorePath
			c := exec.Command("xtrabackup", strings.Split(cmdFlags, " ")...)

			updateLastBackupRecordId(db, backupRecordId)
			err, o := utils.ExecCommand(c)

			utils.CreateFileIfNotExist(incrementalStorePath + "/result.log")
			ioutil.WriteFile(incrementalStorePath+"/result.log", []byte(c.String()+"\n"+o), fs.ModeAppend)

			utils.AddBackupReportToFile(dataDir+"/bk.dat", incrementalStorePath, strategy.Name, timestamp, incrementalTimestamp, err)

			if err != nil {
				incrementalCron.Stop()
				defer utils.PrintRed("[STOP INCREMENTAL WITH ERROR]")
				defer fmt.Println(": " + incrementalStorePath)
			}

			if stopIncremental {
				utils.PrintBlue("[STOP INCREMENTAL]")
				fmt.Println(" with base:" + strconv.FormatInt(timestamp, 10))
			}

			channel <- 1
		})
		incrementalCron.Start()
	}
}
