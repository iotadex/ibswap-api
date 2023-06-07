package model

import (
	"fmt"
	"log"
	"sync"
)

type Coin struct {
	Symbol   string `json:"symbol"`
	ChainID  int64  `json:"chainid"`
	Contract string `json:"contract"`
	Code     string `json:"code"`
	Decimal  int    `json:"decimal"`
	Type     int    `json:"type"`
	Public   int    `json:"public"`
}

type Pool struct {
	ChainID  int64  `json:"chainid"`
	Contract string `json:"contract"`
	Token0   string `json:"token0"`
	Token1   string `json:"token1"`
	FeeRate  int    `json:"fee_rate"`
	Decimal  int    `json:"decimal"`
}

var coinMM map[int64]map[string]*Coin
var coinsM map[int64][]*Coin
var coinsS []*Coin
var coinsMu sync.RWMutex

var poolMM map[int64]map[string]*Pool
var poolsM map[int64][]*Pool
var poolsS []*Pool

func initCoinsAndPools() {
	getCoins()
	getPools()
}

func AddToken(symbol string, chainid int64, contract, code string, decimal, t, public int64) error {
	if _, err := db.Exec("insert into token(`symbol`,`chainid`,`contract`,`code`,`deci`,`type`,`public`) values(?,?,?,?,?,?,?)", symbol, chainid, contract, code, decimal, t, public); err != nil {
		return err
	}
	c := Coin{}
	c.Symbol, c.ChainID, c.Contract, c.Code, c.Decimal, c.Type, c.Public = symbol, chainid, contract, code, int(decimal), int(t), int(public)
	coinsMu.Lock()
	defer coinsMu.Unlock()
	coinMM[c.ChainID][c.Contract] = &c
	coinsM[c.ChainID] = append(coinsM[c.ChainID], &c)
	coinsS = append(coinsS, &c)
	return nil
}

func GetCoin(chainid int64, symbol string) (*Coin, error) {
	coinsMu.RLock()
	defer coinsMu.RUnlock()
	if c, exist := coinMM[chainid][symbol]; exist {
		return c, nil
	} else {
		return c, fmt.Errorf("coin %d:%s is not exist", chainid, symbol)
	}
}

func GetCoinsByChainId(chainid int64) []*Coin {
	coinsMu.RLock()
	defer coinsMu.RUnlock()
	return coinsM[chainid]
}

func GetCoins() []*Coin {
	coinsMu.RLock()
	defer coinsMu.RUnlock()
	return coinsS
}

func GetPool(chainId int64, contract string) (*Pool, error) {
	if c, exist := poolMM[chainId][contract]; exist {
		return c, nil
	} else {
		return c, fmt.Errorf("pool %s+%d is not exist", contract, chainId)
	}
}

func GetPoolsByChainId(chainid int64) []*Pool {
	return poolsM[chainid]
}

func GetPools() []*Pool {
	return poolsS
}

func getCoins() {
	coinMM = make(map[int64]map[string]*Coin)
	coinsM = make(map[int64][]*Coin)
	coinsS = make([]*Coin, 0)
	rows, err := db.Query("select `symbol`,`chainid`,`contract`,`code`,`deci`,`type`,`public` from `token`")
	if err != nil {
		log.Printf("Get coins from db error. %v\n", err)
		return
	}
	for rows.Next() {
		c := Coin{}
		if err = rows.Scan(&c.Symbol, &c.ChainID, &c.Contract, &c.Code, &c.Decimal, &c.Type, &c.Public); err != nil {
			log.Printf("Scan coin from db error. %v\n", err)
			continue
		}
		if _, exist := coinMM[c.ChainID]; !exist {
			coinMM[c.ChainID] = make(map[string]*Coin)
		}
		coinMM[c.ChainID][c.Contract] = &c
		coinsM[c.ChainID] = append(coinsM[c.ChainID], &c)
		coinsS = append(coinsS, &c)
	}
}

func getPools() {
	poolMM = make(map[int64]map[string]*Pool)
	poolsM = make(map[int64][]*Pool)
	poolsS = make([]*Pool, 0)
	rows, err := db.Query("select `chainid`,`contract`,`token0`,`token1`,`fee_rate`,`deci` from `pool`")
	if err != nil {
		log.Printf("Get pools from db error. %v\n", err)
		return
	}
	for rows.Next() {
		p := Pool{}
		if err = rows.Scan(&p.ChainID, &p.Contract, &p.Token0, &p.Token1, &p.FeeRate, &p.Decimal); err != nil {
			log.Printf("Scan pool from db error. %v\n", err)
			continue
		}
		if _, exist := poolMM[p.ChainID]; !exist {
			poolMM[p.ChainID] = make(map[string]*Pool)
		}
		poolMM[p.ChainID][p.Contract] = &p
		poolsM[p.ChainID] = append(poolsM[p.ChainID], &p)
		poolsS = append(poolsS, &p)
	}
}
