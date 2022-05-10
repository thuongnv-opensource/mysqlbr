package service

import (
	"backup-mysql-cli/database"
	"backup-mysql-cli/utils"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/fs"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type ScheduleConfig struct {
	Backup struct {
		DataDir    string             `yaml:"dataBackupDir"`
		Strategies []ScheduleStrategy `yaml:"strategiesBackup"`
	} `yaml:"backup"`
	Mysql MysqlConnectionInfo `yaml:"mysql"`
}

type MysqlConnectionInfo struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
}

type ScheduleStrategy struct {
	Name               string `yaml:"name"`
	Cron               string `yaml:"cron"`
	ExecFlags          string `yaml:"execFlags"`
	IncrementalBackups []struct {
		Every string `yaml:"every"`
		Take  int    `yaml:"take"`
	} `yaml:"incrementalBackups"`
}

func ParseMysqlConnectInfoToXtrabackupFlags(info MysqlConnectionInfo) string {
	result := ""
	if info.User != "" {
		result = result + " --user=" + info.User
	}
	if info.Password != "" {
		result = result + " --password=" + info.Password
	}
	if info.Port != "" {
		result = result + " --port=" + info.Port
	}
	if info.Host != "" {
		result = result + " --host=" + info.Host
	}
	return result
}

func ParseConfigFile(configPath string) ScheduleConfig {
	config := utils.ReadFileInCurrentDirectory(configPath)
	t := ScheduleConfig{}
	err := yaml.Unmarshal([]byte(config), &t)
	fmt.Printf("%+v\n", t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return t
}

func ListBackup(resultFilePath string, dataDir string) {
	for _, record := range utils.ParseBackupReportFile(resultFilePath) {
		if record.HasError {
			utils.PrintRed("[ERROR]")
			fmt.Printf("[%s][%s]:  %s\n", record.Timestamp.Format("01/02/2006 15:04:05"), record.ID, utils.GetBackupRecordPath(dataDir, record))
		} else {
			utils.PrintGreen("[SUCCESS]")
			fmt.Printf("[%s][%s]:  %s\n", record.Timestamp.Format("01/02/2006 15:04:05"), record.ID, utils.GetBackupRecordPath(dataDir, record))
		}
	}
}

func MakeOneSnapshotBackup(config ScheduleConfig) {
	strategyName := "manual"
	dataDir := config.Backup.DataDir
	configPath := utils.ResolveConfigPathFromDataDir(config.Backup.DataDir)

	timestamp := utils.MakeTimestampFromNow()
	utils.CreateFileIfNotExist(configPath)

	backupRecordId := strconv.FormatInt(timestamp, 10)

	storeDataPath := fmt.Sprintf("%s/%s/%s/base", dataDir, strategyName, backupRecordId)

	db := database.GetInstance(config.Mysql.Host, config.Mysql.Port, config.Mysql.User, config.Mysql.Password)
	initDatabase(db)
	updateLastBackupRecordId(db, backupRecordId)

	utils.MakeDirIfNotExist(storeDataPath)
	cmdFlags := fmt.Sprintf("--backup --target-dir=%s", storeDataPath)

	// add execFlags
	cmdFlags = cmdFlags + ParseMysqlConnectInfoToXtrabackupFlags(config.Mysql)

	c := exec.Command("xtrabackup", strings.Split(cmdFlags, " ")...)

	err, o := utils.ExecCommandStreamLog(c)

	utils.CreateFileIfNotExist(storeDataPath + "/result.log")

	ioutil.WriteFile(storeDataPath+"/result.log", []byte(c.String()+"\n"+o), fs.ModeAppend)

	utils.AddBackupReportToFile(configPath, storeDataPath, strategyName, timestamp, 0, err)
}
