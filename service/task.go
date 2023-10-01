package service

import (
	"fmt"
	"ibswap/config"
	"ibswap/gl"
	"ibswap/model"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func Start() {
	volumes24H = make(map[string]*Volumes)
	utc0Reserves = make(map[string]*Reserves)
	currReserves = make(map[string]*Reserves)
	poolStatsM = make(map[string][]model.PoolStat)

	startNft()
	start(3)
	go countVolumes(3)
}

func start(v int8) {
	pools := model.GetPools(v)
	ps := make([]common.Address, 0)
	t0 := make([]common.Address, 0)
	t1 := make([]common.Address, 0)
	count := 0
	for _, p := range pools {
		initPoolStat(p.Contract)
		ps = append(ps, common.HexToAddress(p.Contract))
		t0 = append(t0, common.HexToAddress(p.Token0))
		t1 = append(t1, common.HexToAddress(p.Token1))
		count++
		if count > config.EvmNode.ListenPoolCount {
			StartPool(ps, t0, t1)
			ps = make([]common.Address, 0)
			t0 = make([]common.Address, 0)
			t1 = make([]common.Address, 0)
			count = 0
		}
	}
	if count > 0 {
		StartPool(ps, t0, t1)
	}
}

func initPoolStat(c string) {
	//initial volumes24H
	volumes24H[c] = NewVolumes()
	if ids, vols, err := model.Get24hVolumes(c); err != nil {
		panic(err)
	} else {
		for i := len(vols) - 1; i >= 0; i-- {
			volumes24H[c].append(Volume{amount0: vols[i][0], amount1: vols[i][1], ts: ids[i] * 60})
		}
	}
	//initial utc0Reserves
	utc0Reserves[c] = NewReserves()
	if day, tick, rs, err := model.GetLatestUtc0Reserves(c); err != nil {
		panic(err)
	} else {
		utc0Reserves[c].set(rs[0], rs[1], tick)
		utc0Reserves[c].day = day
	}
	//inital currReserves
	currReserves[c] = NewReserves()
	if rs, tick, err := model.GetLatestReserves(c); err != nil {
		panic(err)
	} else {
		currReserves[c].set(rs[0], rs[1], tick)
	}
}

func initPoolStatReal(c string) {
	//initial volumes24H
	volumes24H[c] = NewVolumes()
	//initial utc0Reserves
	utc0Reserves[c] = NewReserves()
	//inital currReserves
	currReserves[c] = NewReserves()
}

func StartPool(ps, t0, t1 []common.Address) {
	node := NewEvmNode(config.EvmNode.Rpc, config.EvmNode.Wss, config.EvmNode.MaxScanHeight, config.EvmNode.ListenType)
	go dealTickV3(node, ps, t0, t1)
}

func dealTickV3(node *EvmNode, ps, t0, t1 []common.Address) {
	zero := big.NewInt(0)
	chLog, chTick := node.StartListenV3(ps, t0, t1)
	preReserve0, preReserve1 := big.NewInt(0), big.NewInt(0)
	for {
		select {
		case log := <-chLog:
			gl.OutLogger.Error(log)
		case tick := <-chTick:
			key := tick.Address
			if tick.Volume0.Cmp(zero) == 0 && tick.Volume1.Cmp(zero) == 0 {
				_, tick.Tick = currReserves[key].get()
			}

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
			}

			//2. set 24H volumes
			if tick.Volume0.Cmp(zero) > 0 {
				volumes24H[key].append(Volume{amount0: tick.Volume0, amount1: tick.Volume1, ts: tick.Ts})
			}

			//3. store to db
			if err := model.StorePoolVolume(tick.Tx, key, tick.Tick, preReserve0.String(), preReserve1.String(), tick.Volume0.String(), tick.Volume1.String()); err != nil {
				gl.OutLogger.Error("Store pool volume into db error. %s : %v : %v", key, tick, err)
			} else {
				gl.OutLogger.Info("Volume of %s : %d : %s : %s : %s : %s", key, tick.Tick, preReserve0.String(), preReserve1.String(), tick.Volume0.String(), tick.Volume1.String())
			}
		}
	}
}

func startNft() {
	nft := NewEvmNode(config.EvmNode.Rpc, config.EvmNode.Wss, config.EvmNode.MaxScanHeight, config.EvmNode.ListenType)
	go dealNft(nft)
	time.Sleep(time.Second)
}

func dealNft(nft *EvmNode) {
	chLog, chNftToken := nft.StartListenNft(config.EvmNode.Nft, config.EvmNode.Factory, config.EvmNode.InitCode)
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
				if p, err := model.AddPool(nftToken.pool, 3, nftToken.token0, nftToken.token1, nftToken.fee); err != nil {
					gl.OutLogger.Error("Add pool to db error. %v : %v", nftToken, err)
				} else {
					initPoolStatReal(p.Contract)
					StartPool([]common.Address{common.HexToAddress(p.Contract)}, []common.Address{common.HexToAddress(p.Token0)}, []common.Address{common.HexToAddress(p.Token1)})
				}
			}
			if err := model.StoreNftToken(nftToken.tokenId, nftToken.collection, nftToken.user, nftToken.pool, nftToken.token0, nftToken.token1, nftToken.fee); err != nil {
				gl.OutLogger.Error("Store nft token to db error. %v : %v", nftToken, err)
			}
		}
	}
}

func StartFactory() {
	factory := NewEvmNode(config.EvmNode.Rpc, config.EvmNode.Wss, config.EvmNode.MaxScanHeight, config.EvmNode.ListenType)
	go dealFactory(factory)
	time.Sleep(time.Second)
}

func dealFactory(factory *EvmNode) {
	chLog, chPool := factory.StartListenFactory(config.EvmNode.Factory)
	for {
		select {
		case log := <-chLog:
			gl.OutLogger.Error(log)
		case pool := <-chPool:
			gl.OutLogger.Info("Pool had been created. %v", pool)
			if p, err := model.AddPool(pool.Contract, 3, pool.Token0, pool.Token1, pool.FeeRate); err != nil {
				gl.OutLogger.Error("Add pool to db error. %v : %v", pool, err)
				continue
			} else {
				initPoolStatReal(p.Contract)
				StartPool([]common.Address{common.HexToAddress(p.Contract)}, []common.Address{common.HexToAddress(p.Token0)}, []common.Address{common.HexToAddress(p.Token1)})
			}
		}
	}
}

type PoolOverview struct {
	Contract     string `json:"contract"`
	Token0       string `json:"token0"`
	Token1       string `json:"token1"`
	Reserve0     string `json:"reserve0"`
	Reserve1     string `json:"reserve1"`
	CurrTick     int64  `json:"curr_tick"`
	Volume24H0   string `json:"volume24h0"`
	Volume24H1   string `json:"volume24h1"`
	Utc0Reserve0 string `json:"utc0reserve0"`
	Utc0Reserve1 string `json:"utc0reserve1"`
	Utc0Tick     int64  `json:"utc0_tick"`
	Ts           int64  `json:"ts"`
}

func OverviewPools(v int8) []PoolOverview {
	pools := model.GetPools(v)
	ps := make([]PoolOverview, 0)
	for _, p := range pools {
		key := p.Contract
		if _, exist := currReserves[key]; !exist {
			continue
		}
		currReserve, currTick := currReserves[key].get()
		vol24H := volumes24H[key].get24HVolume()
		utc0Reserve, utc0Tick := utc0Reserves[key].get()
		ps = append(ps, PoolOverview{
			Contract:     p.Contract,
			Token0:       p.Token0,
			Token1:       p.Token1,
			Reserve0:     currReserve[0].String(),
			Reserve1:     currReserve[1].String(),
			CurrTick:     currTick,
			Volume24H0:   vol24H.amount0.String(),
			Volume24H1:   vol24H.amount1.String(),
			Utc0Reserve0: utc0Reserve[0].String(),
			Utc0Reserve1: utc0Reserve[1].String(),
			Utc0Tick:     utc0Tick,
			Ts:           vol24H.ts + 86400,
		})
	}
	return ps
}

func OverviewPoolsByContract(contract string) (*PoolOverview, error) {
	p, err := model.GetPool(contract)
	if err != nil {
		return nil, err
	}
	key := contract
	if _, exist := currReserves[key]; !exist {
		return nil, fmt.Errorf("key not exist. %s", key)
	}
	currReserve, currTick := currReserves[key].get()
	vol24H := volumes24H[key].get24HVolume()
	utc0Reserve, utc0Tick := utc0Reserves[key].get()
	return &PoolOverview{
		Contract:     contract,
		Token0:       p.Token0,
		Token1:       p.Token1,
		Reserve0:     currReserve[0].String(),
		Reserve1:     currReserve[1].String(),
		CurrTick:     currTick,
		Volume24H0:   vol24H.amount0.String(),
		Volume24H1:   vol24H.amount1.String(),
		Utc0Reserve0: utc0Reserve[0].String(),
		Utc0Reserve1: utc0Reserve[1].String(),
		Utc0Tick:     utc0Tick,
		Ts:           vol24H.ts + 86400,
	}, nil
}

func StatPoolVolumes(contract string) ([]model.PoolStat, error) {
	currDay := time.Now().Unix() / 86400
	key := contract
	poolStatsMutex.Lock()
	defer poolStatsMutex.Unlock()
	if ps, exist := poolStatsM[key]; exist {
		if len(ps) > 0 {
			if ps[len(ps)-1].Id == currDay-1 {
				return ps, nil
			}
		}
	}
	if ps, err := model.GetPoolStatistic(contract, currDay-config.StatDays); err != nil {
		return nil, err
	} else {
		poolStatsM[key] = ps
		return ps, nil
	}
}

func countVolumes(v int8) {
	preDay := time.Now().Unix() / 86400
	ticker := time.NewTicker(time.Second * 10)
	for range ticker.C {
		currDay := time.Now().Unix() / 86400
		if currDay == preDay {
			continue
		}
		preDay = currDay
		pools := model.GetPools(v)
		for _, p := range pools {
			vol1d := volumes24H[p.Contract].get24HVolume()
			vol7d, err := model.GetNDaysVolumes(p.Contract, currDay-7)
			if err != nil {
				gl.OutLogger.Error("CountPoolVolumes from db error. %s : %v", p.Contract, err)
				continue
			}
			vol7d[0].Add(vol7d[0], vol1d.amount0)
			vol7d[1].Add(vol7d[1], vol1d.amount1)
			currReserve, tick := currReserves[p.Contract].get()
			if err := model.StorePoolStatistic(currDay-1, p.Contract, tick, currReserve[0].String(), currReserve[1].String(), vol1d.amount0.String(), vol1d.amount1.String(), vol7d[0].String(), vol7d[1].String()); err != nil {
				gl.OutLogger.Error("store pool stat into db error. %s : %v", p.Contract, err)
			}
		}
	}
}
