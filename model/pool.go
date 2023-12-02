package model

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"
)

const DefaultTick = 999999999

func StorePoolVolume(tx, contract string, tick int64, r0, r1 string, v0, v1 string) error {
	_, err := db.Exec("insert into volume(`tx`,`contract`,`tick`,`reserve0`,`reserve1`,`vol0`,`vol1`) values(?,?,?,?,?,?,?)", tx, contract, tick, r0, r1, v0, v1)
	if strings.HasPrefix(err.Error(), "Error 1062") {
		return nil
	}
	return err
}

// / returns ts and vol0 and vol1
func Get24hVolumes(contract string) ([]int64, [][2]*big.Int, error) {
	ts := time.Now().AddDate(0, 0, -1).Format(time.DateTime)
	rows, err := db.Query("select `vol0`,`vol1`,`ts` from `volume` where `contract`=? and ts>? order by `ts` desc", contract, ts)
	if err != nil {
		return nil, nil, fmt.Errorf("get 24h volumes from db error. %v", err)
	}
	vols := make([][2]*big.Int, 0)
	tss := make([]int64, 0)
	for rows.Next() {
		var vol0, vol1, ts string
		if err = rows.Scan(&vol0, &vol1, &ts); err != nil {
			return nil, nil, fmt.Errorf("scan 24h volumes from db error. %v", err)
		}
		v0, b0 := new(big.Int).SetString(vol0, 10)
		v1, b1 := new(big.Int).SetString(vol1, 10)
		timestamp, err := time.Parse(time.DateTime, ts)
		if !b0 || !b1 || err != nil {
			return nil, nil, fmt.Errorf("scan 24h volumes from db error. %s : %s : %s", vol0, vol1, ts)
		}
		vols = append(vols, [2]*big.Int{v0, v1})
		tss = append(tss, timestamp.Unix())
	}
	return tss, vols, nil
}

// / returns ts and vol0 and vol1
func Get1DayVolumes(contract string) ([2]*big.Int, error) {
	vols := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	_, vols1d, err := Get24hVolumes(contract)
	if err != nil {
		return vols, err
	}
	for i := range vols1d {
		vols[0].Add(vols[0], vols1d[i][0])
		vols[1].Add(vols[1], vols1d[i][1])
	}
	return vols, nil
}

func GetLatestReserves(c string) ([2]*big.Int, int64, error) {
	row := db.QueryRow("select `reserve0`,`reserve1`,`tick` from `volume` where `contract`=? order by `id` desc limit 1", c)
	var r0, r1 string
	var tick int64
	if err := row.Scan(&r0, &r1, &tick); err != nil {
		if err == sql.ErrNoRows {
			return [2]*big.Int{big.NewInt(0), big.NewInt(0)}, DefaultTick, nil
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

func GetLatestUtc0Reserves(c string) (int64, int64, [2]*big.Int, error) {
	row := db.QueryRow("select `id`,`tick`,`reserve0`,`reserve1` from `pool_stat` where `contract`=? order by `id` desc limit 1", c)
	var r0, r1 string
	var day int64
	var tick int64
	if err := row.Scan(&day, &tick, &r0, &r1); err != nil {
		if err == sql.ErrNoRows {
			return 0, DefaultTick, [2]*big.Int{big.NewInt(0), big.NewInt(0)}, nil
		}
		return day, DefaultTick, [2]*big.Int{}, fmt.Errorf("scan Utc0Reserves from db error. %v", err)
	}
	reserve0, b0 := new(big.Int).SetString(r0, 10)
	reserve1, b1 := new(big.Int).SetString(r1, 10)
	if !b0 || !b1 {
		return day, DefaultTick, [2]*big.Int{}, fmt.Errorf("scan Utc0Reserves from db error. %s : %s", r0, r1)
	}
	return day + 1, tick, [2]*big.Int{reserve0, reserve1}, nil
}

func GetNDaysVolumes(c string, id int64) ([2]*big.Int, error) {
	vols := [2]*big.Int{big.NewInt(0), big.NewInt(0)}
	rows, err := db.Query("select `vol01d`,`vol11d` from `pool_stat` where `contract`=? and id>=?", c, id)
	if err != nil {
		return vols, fmt.Errorf("get n days volumes from db error. %v", err)
	}
	for rows.Next() {
		var vol0, vol1 string
		if err = rows.Scan(&vol0, &vol1); err != nil {
			return vols, fmt.Errorf("scan n days volumes from db error. %v", err)
		}
		v0, b0 := new(big.Int).SetString(vol0, 10)
		v1, b1 := new(big.Int).SetString(vol1, 10)
		if !b0 || !b1 {
			return vols, fmt.Errorf("scan n days volumes from db error. %s : %s", vol0, vol1)
		}
		vols[0].Add(vols[0], v0)
		vols[1].Add(vols[1], v1)
	}
	return vols, nil
}

func StorePoolStatistic(id int64, c string, t int64, r0, r1, v01d, v11d, v07d, v17d string) error {
	_, err := db.Exec("insert into pool_stat(`id`,`contract`,`tick`,`reserve0`,`reserve1`,`vol01d`,`vol11d`,`vol07d`,`vol17d`) values(?,?,?,?,?,?,?,?,?)", id, c, t, r0, r1, v01d, v11d, v07d, v17d)
	if strings.HasPrefix(err.Error(), "Error 1062") {
		return nil
	}
	return err
}

type PoolStat struct {
	Id       int64  `json:"id"`
	Contract string `json:"contract"`
	Tick     int64  `json:"tick"`
	Reserve0 string `json:"reserve0"`
	Reserve1 string `json:"reserve1"`
	Vol01d   string `json:"vol01d"`
	Vol11d   string `json:"vol11d"`
	Vol07d   string `json:"vol07d"`
	Vol17d   string `json:"vol17d"`
}

func GetPoolStatistic(c string, beginDay int64) ([]PoolStat, error) {
	rows, err := db.Query("select `id`,`tick`,`reserve0`,`reserve1`,`vol01d`,`vol11d`,`vol07d`,`vol17d` from `pool_stat` where `contract`=? and id>?", c, beginDay)
	if err != nil {
		return nil, fmt.Errorf("get pool stat from db error. %v", err)
	}
	ps := make([]PoolStat, 0)
	for rows.Next() {
		p := PoolStat{}
		if err = rows.Scan(&p.Id, &p.Tick, &p.Reserve0, &p.Reserve1, &p.Vol01d, &p.Vol11d, &p.Vol07d, &p.Vol17d); err != nil {
			return nil, fmt.Errorf("scan pool stat from db error. %v", err)
		}
		ps = append(ps, p)
	}
	return ps, nil
}
