package service

import (
	"fmt"
	"ibswap/config"
	"ibswap/gl"
	"ibswap/model"
	"math/big"
	"strconv"
	"time"
)

func Start() {
	volumes24H = make(map[string]*Volumes)
	utc0Reserves = make(map[string]*Reserves)
	currReserves = make(map[string]*Reserves)
	poolStatsM = make(map[string][]model.PoolStat)

	start(2)
	start(3)
	startNft()
}

func start(v int8) {
	pools := model.GetPools(v)
	for _, p := range pools {
		StartPool(p)
	}
}

func StartPool(p *model.Pool) {
	key := p.Contract + strconv.FormatInt(p.ChainID, 10)
	//initial volumes24H
	volumes24H[key] = NewVolumes()
	if ids, vols, err := model.Get24hVolumes(p.ChainID, p.Contract); err != nil {
		panic(err)
	} else {
		for i := len(vols) - 1; i >= 0; i-- {
			volumes24H[key].append(Volume{amount0: vols[i][0], amount1: vols[i][1], ts: ids[i] * 60})
		}
	}
	//initial utc0Reserves
	utc0Reserves[key] = &Reserves{}
	if day, rs, err := model.GetLatestUtc0Reserves(p.ChainID, p.Contract); err != nil {
		panic(err)
	} else {
		utc0Reserves[key].set(rs[0], rs[1], 0)
		utc0Reserves[key].day = day
	}
	//inital currReserves
	currReserves[key] = &Reserves{}
	if rs, tick, err := model.GetLatestReserves(p.ChainID, p.Contract); err != nil {
		panic(err)
	} else {
		currReserves[key].set(rs[0], rs[1], tick)
	}

	pool := NewEvmPool(p.ChainID, config.EvmNodes[p.ChainID].Url, p.Contract, p.Token0, p.Token1)
	go dealTick(pool, p.Version)
	time.Sleep(time.Second)
}

func startNft() {
	for chainid, node := range config.EvmNodes {
		nft := NewEvmPool(chainid, node.Url, node.Nft, "0x0", "0x0")
		go dealNft(nft)
		time.Sleep(time.Second)
	}
}

func StartFactory() {
	for chainid, node := range config.EvmNodes {
		factory := NewEvmPool(chainid, node.Url, node.Factory, "0x0", "0x0")
		go dealFactory(factory)
		time.Sleep(time.Second)
	}
}

func dealTick(pool *EvmPool, v int8) {
	zero := big.NewInt(0)
	chLog, chTick := pool.StartListen(v)
	key := pool.contract.Hex() + strconv.FormatInt(pool.chainid, 10)
	preReserve0, preReserve1 := big.NewInt(0), big.NewInt(0)
	for {
		select {
		case log := <-chLog:
			gl.OutLogger.Error(log)
		case tick := <-chTick:
			//1. set current reserves
			if tick.Reserve0.Cmp(zero) != 0 {
				preReserve0, preReserve1 = tick.Reserve0, tick.Reserve1
				currReserves[key].set(preReserve0, preReserve1, tick.Tick)
			}

			//2. update utc0 reserves
			currDay := time.Now().Unix() / 86400
			if utc0Reserves[key].day != currDay {
				utc0Reserves[key].set(preReserve0, preReserve1, tick.Tick)
				utc0Reserves[key].day = currDay

				// count the last days's volume
				v06d, v16d, err := model.Get6DaysVolumes(pool.chainid, pool.contract.Hex())
				if err == nil {
					vol := volumes24H[key].get24HVolume()
					v06d.Add(v06d, vol.amount0)
					v16d.Add(v16d, vol.amount1)
					if err := model.StorePoolStatistic(currDay-1, pool.chainid, pool.contract.Hex(), preReserve0.String(), preReserve1.String(), vol.amount0.String(), vol.amount1.String(), v06d.String(), v16d.String()); err != nil {
						gl.OutLogger.Error("store pool stat into db error. %s : %v", key, err)
					}
				} else {
					gl.OutLogger.Error("Get6DaysVolumes from db error. %s : %v", key, err)
				}
			}

			//2. set 24H volumes
			if tick.Volume0.Cmp(zero) > 0 {
				volumes24H[key].append(Volume{amount0: tick.Volume0, amount1: tick.Volume1, ts: tick.Ts})
			}

			id := time.Now().Unix() / 60
			//3. store to db
			if err := model.StorePoolVolume(id, pool.chainid, pool.contract.Hex(), tick.Tick, preReserve0.String(), preReserve1.String(), tick.Volume0.String(), tick.Volume1.String()); err != nil {
				gl.OutLogger.Error("Store pool volume into db error. %s : %v : %v", key, tick, err)
			} else {
				gl.OutLogger.Info("Volume of %s : %d : %s : %s : %s : %s", key, tick.Tick, preReserve0.String(), preReserve1.String(), tick.Volume0.String(), tick.Volume1.String())
			}
		}
	}
}

func dealNft(nft *EvmPool) {
	chLog, chNftToken := nft.StartListenNFT(config.EvmNodes[nft.chainid].Factory, config.EvmNodes[nft.chainid].InitCode)
	for {
		select {
		case log := <-chLog:
			gl.OutLogger.Error(log)
		case nftToken := <-chNftToken:
			gl.OutLogger.Info("NFT token record. %v", nftToken)
			if nftToken.direction == -1 {
				if err := model.DeleteNftToken(nftToken.tokenId, nftToken.collection); err != nil {
					gl.OutLogger.Error("Delete nft token error. %v : %v", nftToken, err)
				}
				continue
			}
			//Get the pool's contract
			if p := model.GetPoolByTokensAndFee(nftToken.token0, nftToken.token1, nftToken.fee); p == nil {
				if err := model.AddPool(nft.chainid, nftToken.pool, 3, nftToken.token0, nftToken.token1, nftToken.fee); err != nil {
					gl.OutLogger.Error("Add pool to db error. %v : %v", nftToken, err)
				}
			}
			if err := model.StoreNftToken(nftToken.tokenId, nftToken.collection, nftToken.user, nftToken.pool, nftToken.token0, nftToken.token1, nftToken.fee); err != nil {
				gl.OutLogger.Error("Store nft token to db error. %v : %v", nftToken, err)
			}
		}
	}
}

func dealFactory(factory *EvmPool) {
	chLog, chPool := factory.StartListenFactory()
	for {
		select {
		case log := <-chLog:
			gl.OutLogger.Error(log)
		case pool := <-chPool:
			gl.OutLogger.Info("Pool had been created. %v", pool)
			if err := model.AddPool(factory.chainid, pool.Contract, 3, pool.Token0, pool.Token1, pool.FeeRate); err != nil {
				gl.OutLogger.Error("Add pool to db error. %v : %v", pool, err)
				continue
			}
			if p, err := model.GetPool(factory.chainid, pool.Contract); err != nil {
				gl.OutLogger.Error("Get pool from cache error not exist. %v : %v", pool, err)
			} else {
				StartPool(p)
			}
		}
	}
}

type PoolOverview struct {
	ChainID      int64  `json:"chainid"`
	Contract     string `json:"contract"`
	Token0       string `json:"token0"`
	Token1       string `json:"token1"`
	Reserve0     string `json:"reserve0"`
	Reserve1     string `json:"reserve1"`
	Tick         int64  `json:"tick"`
	Volume24H0   string `json:"volume24h0"`
	Volume24H1   string `json:"volume24h1"`
	Utc0Reserve0 string `json:"utc0reserve0"`
	Utc0Reserve1 string `json:"utc0reserve1"`
	Ts           int64  `json:"ts"`
}

func OverviewPoolsByChainid(chainid int64, v int8) []PoolOverview {
	pools := model.GetPoolsByChainId(chainid, v)
	ps := make([]PoolOverview, 0)
	for _, p := range pools {
		key := p.Contract + strconv.FormatInt(chainid, 10)
		if _, exist := currReserves[key]; !exist {
			continue
		}
		currReserve, currTick := currReserves[key].get()
		vol24H := volumes24H[key].get24HVolume()
		utc0Reserve, _ := utc0Reserves[key].get()
		ps = append(ps, PoolOverview{
			ChainID:      chainid,
			Contract:     p.Contract,
			Token0:       p.Token0,
			Token1:       p.Token1,
			Reserve0:     currReserve[0].String(),
			Reserve1:     currReserve[1].String(),
			Tick:         currTick,
			Volume24H0:   vol24H.amount0.String(),
			Volume24H1:   vol24H.amount1.String(),
			Utc0Reserve0: utc0Reserve[0].String(),
			Utc0Reserve1: utc0Reserve[0].String(),
			Ts:           vol24H.ts + 86400,
		})
	}
	return ps
}

func OverviewPoolsByChainidAndContract(chainid int64, contract string) (*PoolOverview, error) {
	p, err := model.GetPool(chainid, contract)
	if err != nil {
		return nil, err
	}
	key := contract + strconv.FormatInt(chainid, 10)
	if _, exist := currReserves[key]; !exist {
		return nil, fmt.Errorf("key not exist. %s", key)
	}
	currReserve, currTick := currReserves[key].get()
	vol24H := volumes24H[key].get24HVolume()
	utc0Reserve, _ := utc0Reserves[key].get()
	return &PoolOverview{
		ChainID:      chainid,
		Contract:     contract,
		Token0:       p.Token0,
		Token1:       p.Token1,
		Reserve0:     currReserve[0].String(),
		Reserve1:     currReserve[1].String(),
		Tick:         currTick,
		Volume24H0:   vol24H.amount0.String(),
		Volume24H1:   vol24H.amount1.String(),
		Utc0Reserve0: utc0Reserve[0].String(),
		Utc0Reserve1: utc0Reserve[1].String(),
		Ts:           vol24H.ts + 86400,
	}, nil
}

func CountPoolVolumes(chainid int64, contract string) ([]model.PoolStat, error) {
	currDay := time.Now().Unix() / 86400
	key := contract + strconv.FormatInt(chainid, 10)
	poolStatsMutex.Lock()
	defer poolStatsMutex.Unlock()
	if ps, exist := poolStatsM[key]; exist {
		if len(ps) > 0 {
			if ps[len(ps)-1].Id == currDay-1 {
				return ps, nil
			}
		}
	}
	if ps, err := model.GetPoolStatistic(chainid, contract, currDay-config.StatDays); err != nil {
		return nil, err
	} else {
		poolStatsM[key] = ps
		return ps, nil
	}
}
