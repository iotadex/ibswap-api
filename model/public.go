package model

import (
	"fmt"
	"log"
	"sync"
)

type Coin struct {
	Symbol   string `json:"symbol"`
	Contract string `json:"contract"`
	Code     string `json:"code"`
	Decimal  int    `json:"decimal"`
	Type     int    `json:"type"`
	Public   int    `json:"public"`
}

type Pool struct {
	Contract string `json:"contract"`
	Version  int8   `json:"version"`
	Token0   string `json:"token0"`
	Token1   string `json:"token1"`
	FeeRate  int    `json:"fee_rate"`
	Decimal  int    `json:"decimal"`
	State    int    `json:"state"`
}

var coinM map[string]*Coin
var coins []*Coin
var coinsMu sync.RWMutex

var poolMV2 map[string]*Pool
var poolsV2 []*Pool

var poolMV3 map[string]*Pool
var poolsV3 []*Pool

var poolMMM map[string]map[string]map[int]*Pool //token0->token1->fee->Pool

func initCoinsAndPools() {
	getCoins()
	getPools()
}

func AddToken(symbol string, contract, code string, decimal, t, public int64) error {
	if _, err := db.Exec("insert into `token`(`symbol`,`contract`,`code`,`deci`,`type`,`public`) values(?,?,?,?,?,?)", symbol, contract, code, decimal, t, public); err != nil {
		return err
	}
	c := Coin{}
	c.Symbol, c.Contract, c.Code, c.Decimal, c.Type, c.Public = symbol, contract, code, int(decimal), int(t), int(public)
	coinsMu.Lock()
	defer coinsMu.Unlock()
	coinM[c.Contract] = &c
	coins = append(coins, &c)
	return nil
}

func AddPool(contract string, version int64, token0, token1 string, feeRate int) (*Pool, error) {
	if _, err := db.Exec("insert into `pool`(`contract`,`version`,`token0`,`token1`,`fee_rate`) values(?,?,?,?,?)", contract, version, token0, token1, feeRate); err != nil {
		return nil, err
	}
	p := Pool{
		Contract: contract,
		Version:  int8(version),
		Token0:   token0,
		Token1:   token1,
		FeeRate:  feeRate,
		Decimal:  18,
	}
	addPool(&p)
	return &p, nil
}

func ChangePoolState(contract string, state int) error {
	if _, err := db.Exec("update `pool` set state=? where contract=?", state, contract); err != nil {
		return err
	}

	if p := poolMV3[contract]; p != nil {
		p.State = state
	}
	return nil
}

func ChangeTokenPublic(contract string, public int) error {
	if _, err := db.Exec("update `token` set public=? where contract=?", public, contract); err != nil {
		return err
	}

	if c := coinM[contract]; c != nil {
		c.Public = public
	}
	return nil
}

func GetCoin(symbol string) (*Coin, error) {
	coinsMu.RLock()
	defer coinsMu.RUnlock()
	if c, exist := coinM[symbol]; exist {
		return c, nil
	} else {
		return c, fmt.Errorf("coin %s is not exist", symbol)
	}
}

func GetCoins() []*Coin {
	coinsMu.RLock()
	defer coinsMu.RUnlock()
	return coins
}

func GetPool(contract string) (*Pool, error) {
	if p, exist := poolMV2[contract]; exist {
		return p, nil
	}
	if p, exist := poolMV3[contract]; exist {
		return p, nil
	}
	return nil, nil
}

func GetPoolByTokensAndFee(token0, token1 string, fee int) *Pool {
	if p, exist := poolMMM[token0][token1][fee]; exist {
		return p
	}
	return nil
}

func GetPools(v int8) []*Pool {
	if v == 2 {
		return poolsV2
	}
	return poolsV3
}

func getCoins() {
	coinM = make(map[string]*Coin)
	coins = make([]*Coin, 0)
	rows, err := db.Query("select `symbol`,`contract`,`code`,`deci`,`type`,`public` from `token`")
	if err != nil {
		log.Printf("Get coins from db error. %v\n", err)
		return
	}
	for rows.Next() {
		c := Coin{}
		if err = rows.Scan(&c.Symbol, &c.Contract, &c.Code, &c.Decimal, &c.Type, &c.Public); err != nil {
			log.Printf("Scan coin from db error. %v\n", err)
			continue
		}
		coinM[c.Contract] = &c
		coins = append(coins, &c)
	}
}

func getPools() {
	poolMV2 = make(map[string]*Pool)
	poolsV2 = make([]*Pool, 0)
	poolMV3 = make(map[string]*Pool)
	poolsV3 = make([]*Pool, 0)
	poolMMM = make(map[string]map[string]map[int]*Pool)
	rows, err := db.Query("select `contract`,`version`,`token0`,`token1`,`fee_rate`,`deci`,`state` from `pool`")
	if err != nil {
		log.Printf("Get pools from db error. %v\n", err)
		return
	}
	for rows.Next() {
		p := Pool{}
		if err = rows.Scan(&p.Contract, &p.Version, &p.Token0, &p.Token1, &p.FeeRate, &p.Decimal, &p.State); err != nil {
			log.Printf("Scan pool from db error. %v\n", err)
			continue
		}
		addPool(&p)
	}
}

func addPool(p *Pool) {
	if p.Version == 2 {
		poolMV2[p.Contract] = p
		poolsV2 = append(poolsV2, p)
	} else if p.Version == 3 {
		poolMV3[p.Contract] = p
		poolsV3 = append(poolsV3, p)
	}

	if _, exist := poolMMM[p.Token0]; !exist {
		poolMMM[p.Token0] = make(map[string]map[int]*Pool)
	}
	if _, exist := poolMMM[p.Token0][p.Token1]; !exist {
		poolMMM[p.Token0][p.Token1] = make(map[int]*Pool)
	}
	poolMMM[p.Token0][p.Token1][p.FeeRate] = p
}
