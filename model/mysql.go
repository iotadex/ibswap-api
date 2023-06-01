package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"ibswap/config"
	"log"
	"os"

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

func LoadPoolFromJosn() {
	file, err := os.Open("pool.json")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	pools := make([]Pool, 0)
	if err = json.NewDecoder(file).Decode(&pools); err != nil {
		log.Panic(err)
	}
	for _, p := range pools {
		if _, err := db.Exec("insert into `pool`(`chainid`,`contract`,`token0`,`token1`,`fee_rate`,`deci`) values(?,?,?,?,?,?)", p.ChainID, p.Contract, p.Token0, p.Token1, p.FeeRate, p.Decimal); err != nil {
			log.Panic(err)
		}
	}
}
