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
	Version  int8
	Token0   string `json:"token0"`
	Token1   string `json:"token1"`
	FeeRate  int    `json:"fee_rate"`
	Decimal  int    `json:"decimal"`
}

var coinMM map[int64]map[string]*Coin
var coinsM map[int64][]*Coin
var coinsS []*Coin
var coinsMu sync.RWMutex

var poolMMV2 map[int64]map[string]*Pool
var poolsMV2 map[int64][]*Pool
var poolsSV2 []*Pool

var poolMMV3 map[int64]map[string]*Pool
var poolsMV3 map[int64][]*Pool
var poolsSV3 []*Pool
var poolMMM map[string]map[string]map[int]*Pool //token0->token1->fee->Pool

func initCoinsAndPools() {
	getCoins()
	getPools()
}

func AddToken(symbol string, chainid int64, contract, code string, decimal, t, public int64) error {
	if _, err := db.Exec("insert into `token`(`symbol`,`chainid`,`contract`,`code`,`deci`,`type`,`public`) values(?,?,?,?,?,?,?)", symbol, chainid, contract, code, decimal, t, public); err != nil {
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

func AddPool(chainid int64, contract string, version int64, token0, token1 string, feeRate int) (*Pool, error) {
	if _, err := db.Exec("insert into `pool`(`chainid`,`contract`,`version`,`token0`,`token1`,`fee_rate`) values(?,?,?,?,?,?)", chainid, contract, version, token0, token1, feeRate); err != nil {
		return nil, err
	}
	p := Pool{
		ChainID:  chainid,
		Contract: contract,
		Version:  int8(version),
		Token0:   token0,
		Token1:   token1,
		FeeRate:  feeRate,
		Decimal:  18,
	}
	poolMMV3[chainid][contract] = &p
	poolsMV3[chainid] = append(poolsMV3[chainid], &p)
	poolsSV3 = append(poolsSV3, &p)
	if _, exist := poolMMM[p.Token0]; !exist {
		poolMMM[p.Token0] = make(map[string]map[int]*Pool)
	}
	if _, exist := poolMMM[p.Token0][p.Token1]; !exist {
		poolMMM[p.Token0][p.Token1] = make(map[int]*Pool)
	}
	poolMMM[p.Token0][p.Token1][p.FeeRate] = &p
	return &p, nil
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
	if p, exist := poolMMV2[chainId][contract]; exist {
		return p, nil
	}
	if p, exist := poolMMV3[chainId][contract]; exist {
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

func GetPoolsByChainId(chainid int64, v int8) []*Pool {
	if v == 2 {
		return poolsMV2[chainid]
	}
	return poolsMV3[chainid]
}

func GetPools(v int8) []*Pool {
	if v == 2 {
		return poolsSV2
	}
	return poolsSV3
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
	poolMMV2 = make(map[int64]map[string]*Pool)
	poolsMV2 = make(map[int64][]*Pool)
	poolsSV2 = make([]*Pool, 0)
	poolMMV3 = make(map[int64]map[string]*Pool)
	poolsMV3 = make(map[int64][]*Pool)
	poolsSV3 = make([]*Pool, 0)
	poolMMM = make(map[string]map[string]map[int]*Pool)
	rows, err := db.Query("select `chainid`,`contract`,`version`,`token0`,`token1`,`fee_rate`,`deci` from `pool`")
	if err != nil {
		log.Printf("Get pools from db error. %v\n", err)
		return
	}
	for rows.Next() {
		p := Pool{}
		if err = rows.Scan(&p.ChainID, &p.Contract, &p.Version, &p.Token0, &p.Token1, &p.FeeRate, &p.Decimal); err != nil {
			log.Printf("Scan pool from db error. %v\n", err)
			continue
		}
		if p.Version == 2 {
			if _, exist := poolMMV2[p.ChainID]; !exist {
				poolMMV2[p.ChainID] = make(map[string]*Pool)
			}
			poolMMV2[p.ChainID][p.Contract] = &p
			poolsMV2[p.ChainID] = append(poolsMV2[p.ChainID], &p)
			poolsSV2 = append(poolsSV2, &p)
		} else if p.Version == 3 {
			if _, exist := poolMMV3[p.ChainID]; !exist {
				poolMMV3[p.ChainID] = make(map[string]*Pool)
			}
			poolMMV3[p.ChainID][p.Contract] = &p
			poolsMV3[p.ChainID] = append(poolsMV3[p.ChainID], &p)
			poolsSV3 = append(poolsSV3, &p)
			if _, exist := poolMMM[p.Token0]; !exist {
				poolMMM[p.Token0] = make(map[string]map[int]*Pool)
			}
			if _, exist := poolMMM[p.Token0][p.Token1]; !exist {
				poolMMM[p.Token0][p.Token1] = make(map[int]*Pool)
			}
			poolMMM[p.Token0][p.Token1][p.FeeRate] = &p
		}
	}
}
