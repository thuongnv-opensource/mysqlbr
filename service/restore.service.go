package service

import (
	"backup-mysql-cli/utils"
	"fmt"
	"log"
	"os/exec"
	"sort"
	"strings"
)

func PrepareBackup(dataDir string, id string) string {
	var sourcePath = utils.ResolveConfigPathFromDataDir(dataDir)

	if strings.Contains(sourcePath, "@") {
		// ssh source
		err, f := RsyncCopyFile(sourcePath, "", "")
		if err != nil {
			log.Fatal(err)
		}
		sourcePath = f
	}
	var baseId string
	var parse = utils.ParseBackupReportFile(sourcePath)

	sort.SliceStable(parse, func(i, j int) bool {
		return parse[i].Timestamp.UnixMilli() < parse[j].Timestamp.UnixMilli()
	})

	for _, record := range parse {
		if record.ID == id {
			baseId = record.BaseStr
		}
	}
	result := make([]utils.BackupRecord, 0)
	for _, record := range parse {
		if record.ID == id {
			result = append(result, record)
			break
		}
		if record.BaseStr == baseId {
			result = append(result, record)
		}
	}

	var baseDir string
	for index, record := range result {
		if record.HasError {
			log.Fatal("Backup " + record.ID + " has error")
		}
		err, r := RsyncCopyDir(utils.GetBackupRecordPath(dataDir, record)+"/", "", "")
		if err != nil {
			log.Fatal(err)
		}
		if record.Incremental == nil {
			// base
			if len(result) == 1 {
				c := exec.Command("xtrabackup", "--prepare", "--target-dir", r)
				err, _ := utils.ExecCommandStreamLog(c)
				if err != nil {
					log.Fatal(err)
				}
				baseDir = r
			} else {
				c := exec.Command("xtrabackup", "--prepare", "--apply-log-only", "--target-dir", r)
				err, _ := utils.ExecCommandStreamLog(c)
				if err != nil {
					log.Fatal(err)
				}
				baseDir = r
			}

		}

		if record.Incremental != nil && index < len(result)-1 {
			c := exec.Command("xtrabackup", "--prepare", "--apply-log-only", "--target-dir", baseDir, "--incremental-dir", r)
			err, _ := utils.ExecCommandStreamLog(c)
			if err != nil {
				log.Fatal(err)
			}
		}

		/*
		*	--apply-log-only should be used when merging all incrementals except the last one.
		 */
		if record.Incremental != nil && index == len(result)-1 {
			c := exec.Command("xtrabackup", "--prepare", "--target-dir", baseDir, "--incremental-dir", r)
			err, _ := utils.ExecCommandStreamLog(c)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Println("=============================================")
	fmt.Println("To continue exec command:")
	fmt.Println("rm -rf /var/lib/mysql/* # remove data content")
	fmt.Println("rm -rf /var/log/mysql/* # remove log content")
	fmt.Println("xtrabackup --copy-back --target-dir=" + baseDir)
	fmt.Println("chown -R mysql:mysql /var/lib/mysql")
	fmt.Println("systemctl restart mysql")

	return baseDir
}
