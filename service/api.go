package service

import (
	"fmt"
	"ibdex/config"
	"ibdex/gl"
	"ibdex/model"
	"time"
)

func GetEthPrice() float64 {
	return currentEthPrice.Get()
}

func GetTokenUsdPrice(contract string) float64 {
	if tokenPricesAsUsd[contract] != nil {
		return tokenPricesAsUsd[contract].Get()
	}
	return 0
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
