package service

import (
	"ibswap/model"
	"math/big"
	"sync"
	"time"
)

type Reserves struct {
	day  int64
	data [2]*big.Int
	mu   sync.RWMutex
}

func (r *Reserves) set(r0, r1 *big.Int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[0] = r0
	r.data[1] = r1
}

func (r *Reserves) get() [2]*big.Int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var data [2]*big.Int = r.data
	return data
}

var utc0Reserves map[string]*Reserves
var currReserves map[string]*Reserves

type Volumes struct {
	data   []Volume
	vol24H Volume
	mu     sync.Mutex
}

type Volume struct {
	amount0 *big.Int
	amount1 *big.Int
	ts      int64
}

func NewVolumes() *Volumes {
	return &Volumes{
		data: make([]Volume, 0),
		vol24H: Volume{
			amount0: big.NewInt(0),
			amount1: big.NewInt(0),
		},
	}
}

func (v *Volumes) append(vol Volume) {
	v.get24HVolume()
	v.mu.Lock()
	defer v.mu.Unlock()
	v.data = append(v.data, vol)
	v.vol24H.amount0.Add(v.vol24H.amount0, vol.amount0)
	v.vol24H.amount1.Add(v.vol24H.amount1, vol.amount1)
}

func (v *Volumes) get24HVolume() Volume {
	v.mu.Lock()
	defer v.mu.Unlock()
	before24H := time.Now().Unix() - 86400
	for len(v.data) > 0 {
		if v.data[0].ts < before24H {
			v.vol24H.amount0.Sub(v.vol24H.amount0, v.data[0].amount0)
			v.vol24H.amount1.Sub(v.vol24H.amount1, v.data[0].amount1)
			v.data = v.data[1:]
			continue
		}
		break
	}
	v.vol24H.ts = before24H
	return v.vol24H
}

var volumes24H map[string]*Volumes

var poolStatsM map[string][]model.PoolStat
var poolStatsMutex sync.Mutex
