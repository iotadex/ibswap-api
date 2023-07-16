package config

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type db struct {
	Host   string `json:"host"`
	Port   string `json:"port"`
	DbName string `json:"dbname"`
	Usr    string `json:"usr"`
	Pwd    string `json:"pwd"`
}

type evmNode struct {
	Rpc             string `json:"rpc"`
	Wss             string `json:"wss"`
	MaxScanHeight   uint64 `json:"max_scan_height"`
	ListenType      int    `json:"listen_type"`
	ListenPoolCount int    `json:"listen_pool_count"`
	Nft             string `json:"nft"`
	Factory         string `json:"factory"`
	InitCode        string `json:"init_code"`
}

var (
	Db       db
	HttpPort int
	EvmNode  evmNode //chainid : url
	ScanTime time.Duration
	StatDays int64
)

// Load load config file
func init() {
	file, err := os.Open("config/config.json")
	if err != nil {
		log.Panic(err)
	}
	defer file.Close()
	type Config struct {
		HttpPort int           `json:"http_port"`
		Db       db            `json:"db"`
		EvmNode  evmNode       `json:"evm_node"`
		ScanTime time.Duration `json:"scan_time"`
		StatDays int64         `json:"stat_days"`
	}
	all := &Config{}
	if err = json.NewDecoder(file).Decode(all); err != nil {
		log.Panic(err)
	}
	Db = all.Db
	HttpPort = all.HttpPort
	EvmNode = all.EvmNode
	ScanTime = all.ScanTime
	StatDays = all.StatDays
}
