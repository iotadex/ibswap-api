package model

import (
	"database/sql"
	"fmt"
	"ibdex/config"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func ConnectToMysql() {
	usr, pwd := config.Db.Usr, config.Db.Pwd
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", usr, pwd, config.Db.Host, config.Db.Port, config.Db.DbName))
	if err != nil {
		log.Panic(err)
	}

	if err = db.Ping(); nil != err {
		log.Panic("Connect to Mysql error : " + err.Error())
	}

	initCoinsAndPools()
}

func Ping() error {
	if db == nil {
		return fmt.Errorf("mysql connection is nil")
	}
	if err := db.Ping(); nil != err {
		return fmt.Errorf("connect to Mysql error : %v", err)
	}
	return nil
}
