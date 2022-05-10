package main

import (
	"backup-mysql-cli/cmd"
)

func main() {
	cmd.Execute()
	//service.ScheduleBackup("schedule:\n  dataDir: 'backup'\n  strategies:\n    - name: hang-ngay\n      ExecFlags: \"\"\n      cron: \"* * * * * *\"\n      incrementalBackups:\n        - every: \"10s\"\n          take: 5\n    - name: hang-tuan\n      ExecFlags: \"\"\n      cron: \"* * * * *\"\n      incrementalBackups:\n        - every: \"10s\"\n          take: 5\n")
	//service.ListBackup("schedule:\n  dataDir: 'backup'\n  strategies:\n    - name: hang-ngay\n      ExecFlags: \"\"\n      cron: \"* * * * * *\"\n      incrementalBackups:\n        - every: \"10s\"\n          take: 5\n    - name: hang-tuan\n      ExecFlags: \"\"\n      cron: \"* * * * *\"\n      incrementalBackups:\n        - every: \"10s\"\n          take: 5\n")

	//fmt.Println(utils.ExecCommand(exec.Command("xtrabackup","--backup","--target-dir=backup/hang-ngay/1652174507000/base")))

	//fmt.Println(service.RsyncCopyDir("vagrant@192.168.56.13:/home/vagrant/backup", "", "test123"))

	//service.PrepareBackup("vagrant@192.168.56.13:/home/vagrant/backup", "1652238641001")

}
