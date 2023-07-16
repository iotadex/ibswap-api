package model

import (
	"database/sql"
	"fmt"
	"math/big"
	"time"
)

func StorePoolVolume(id int64, contract string, tick int64, r0, r1 string, v0, v1 string) error {
	_, err := db.Exec("insert into volume(`id`,`contract`,`tick`,`reserve0`,`reserve1`,`vol0`,`vol1`) values(?,?,?,?,?,?,?)", id, contract, tick, r0, r1, v0, v1)
	return err
}

func Get24hVolumes(contract string) ([]int64, [][2]*big.Int, error) {
	id := time.Now().Unix()/60 - 1440
	rows, err := db.Query("select `id`,`vol0`,`vol1` from `volume` where `contract`=? and id>? order by `id` desc", contract, id)
	if err != nil {
		return nil, nil, fmt.Errorf("get 24h volumes from db error. %v", err)
	}
	vols := make([][2]*big.Int, 0)
	ids := make([]int64, 0)
	for rows.Next() {
		var vol0, vol1 string
		if err = rows.Scan(&id, &vol0, &vol1); err != nil {
			return nil, nil, fmt.Errorf("scan 24h volumes from db error. %v", err)
		}
		v0, b0 := new(big.Int).SetString(vol0, 10)
		v1, b1 := new(big.Int).SetString(vol1, 10)
		if !b0 || !b1 {
			return nil, nil, fmt.Errorf("scan 24h volumes from db error. %s : %s", vol0, vol1)
		}
		vols = append(vols, [2]*big.Int{v0, v1})
		ids = append(ids, id)
	}
	return ids, vols, nil
}

func GetLatestReserves(c string) ([2]*big.Int, int64, error) {
	row := db.QueryRow("select `reserve0`,`reserve1`,`tick` from `volume` where `contract`=? order by `id` desc limit 1", c)
	var r0, r1 string
	var tick int64
	if err := row.Scan(&r0, &r1, &tick); err != nil {
		if err == sql.ErrNoRows {
			return [2]*big.Int{big.NewInt(0), big.NewInt(0)}, 0, nil
		}
		return [2]*big.Int{}, 0, fmt.Errorf("scan LatestReserves from db error. %v", err)
	}
	reserve0, b0 := new(big.Int).SetString(r0, 10)
	reserve1, b1 := new(big.Int).SetString(r1, 10)
	if !b0 || !b1 {
		return [2]*big.Int{}, 0, fmt.Errorf("scan Utc0Reserves from db error. %s : %s", r0, r1)
	}
	return [2]*big.Int{reserve0, reserve1}, tick, nil
}

func GetLatestUtc0Reserves(c string) (int64, [2]*big.Int, error) {
	row := db.QueryRow("select `id`,`reserve0`,`reserve1` from `pool_stat` where `contract`=? order by `id` desc limit 1", c)
	var r0, r1 string
	var day int64
	if err := row.Scan(&day, &r0, &r1); err != nil {
		if err == sql.ErrNoRows {
			return 0, [2]*big.Int{big.NewInt(0), big.NewInt(0)}, nil
		}
		return day, [2]*big.Int{}, fmt.Errorf("scan Utc0Reserves from db error. %v", err)
	}
	reserve0, b0 := new(big.Int).SetString(r0, 10)
	reserve1, b1 := new(big.Int).SetString(r1, 10)
	if !b0 || !b1 {
		return day, [2]*big.Int{}, fmt.Errorf("scan Utc0Reserves from db error. %s : %s", r0, r1)
	}
	return day + 1, [2]*big.Int{reserve0, reserve1}, nil
}

func Get6DaysVolumes(c string) (*big.Int, *big.Int, error) {
	rows, err := db.Query("select `vol01d`,`vol11d` from `pool_stat` where `contract`=? order by `id` desc limit 6", c)
	if err != nil {
		return nil, nil, fmt.Errorf("get 6 days volumes from db error. %v", err)
	}
	v06d, v16d := big.NewInt(0), big.NewInt(0)
	for rows.Next() {
		var vol0, vol1 string
		if err = rows.Scan(&vol0, &vol1); err != nil {
			return nil, nil, fmt.Errorf("scan 6 days volumes from db error. %v", err)
		}
		v0, b0 := new(big.Int).SetString(vol0, 10)
		v1, b1 := new(big.Int).SetString(vol1, 10)
		if !b0 || !b1 {
			return nil, nil, fmt.Errorf("scan 6 days volumes from db error. %s : %s", vol0, vol1)
		}
		v06d.Add(v06d, v0)
		v16d.Add(v16d, v1)
	}
	return v06d, v16d, nil
}

func StorePoolStatistic(id int64, c, r0, r1, v01d, v11d, v07d, v17d string) error {
	_, err := db.Exec("insert into pool_stat(`id`,`contract`,`reserve0`,`reserve1`,`vol01d`,`vol11d`,`vol07d`,`vol17d`) values(?,?,?,?,?,?,?,?)", id, c, r0, r1, v01d, v11d, v07d, v17d)
	return err
}

type PoolStat struct {
	Id       int64  `json:"id"`
	Contract string `json:"contract"`
	Reserve0 string `json:"reserve0"`
	Reserve1 string `json:"reserve1"`
	Vol01d   string `json:"vol01d"`
	Vol11d   string `json:"vol11d"`
	Vol07d   string `json:"vol07d"`
	Vol17d   string `json:"vol17d"`
}

func GetPoolStatistic(c string, beginDay int64) ([]PoolStat, error) {
	rows, err := db.Query("select `id`,`reserve0`,`reserve1`,`vol01d`,`vol11d`,`vol07d`,`vol17d` from `pool_stat` where `contract`=? and id>?", c, beginDay)
	if err != nil {
		return nil, fmt.Errorf("get pool stat from db error. %v", err)
	}
	ps := make([]PoolStat, 0)
	for rows.Next() {
		p := PoolStat{}
		if err = rows.Scan(&p.Id, &p.Reserve0, &p.Reserve1, &p.Vol01d, &p.Vol11d, &p.Vol07d, &p.Vol17d); err != nil {
			return nil, fmt.Errorf("scan pool stat from db error. %v", err)
		}
		ps = append(ps, p)
	}
	return ps, nil
}
