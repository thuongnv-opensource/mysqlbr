package database

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"sync"
)

func connect(host string, port string, user string, password string) *gorm.DB {

	var dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/", user, password, host, port)
	fmt.Println("[CONNECT STRING]: " + dsn)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic("failed to connect database")
	}

	return db
}

var lock = &sync.Mutex{}
var db *gorm.DB

func GetInstance(host string, port string, user string, password string) *gorm.DB {
	if db == nil {
		lock.Lock()
		defer lock.Unlock()
		if db == nil {
			db = connect(host, port, user, password)
		}
	}
	return db
}
