package service

import (
	"fmt"
	"ibdex/config"
	"ibdex/contracts"
	"ibdex/gl"
	"ibdex/model"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Start() {
	volumes24H = make(map[string]*Volumes)
	utc0Reserves = make(map[string]*Reserves)
	currReserves = make(map[string]*Reserves)
	poolStatsM = make(map[string][]model.PoolStat)

	startNft()
	start(3)
	go countVolumes(3)

	RealEthPrice()
	RealProtocolFees()
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
	if tss, vols, err := model.Get24hVolumes(c); err != nil {
		panic(err)
	} else {
		for i := len(vols) - 1; i >= 0; i-- {
			volumes24H[c].append(Volume{amount0: vols[i][0], amount1: vols[i][1], ts: tss[i]})
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
				cr, ct := currReserves[key].get()
				utc0Reserves[key].set(cr[0], cr[1], ct)
				utc0Reserves[key].day = currDay
			}

			//2. set 24H volumes
			if tick.Volume0.Cmp(zero) > 0 || tick.Volume1.Cmp(zero) > 0 {
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
	c, err := ethclient.Dial(config.EvmNode.Rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}
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
			addCoin(c, nftToken.token0)
			addCoin(c, nftToken.token1)
		}
	}
}

func StartFactory() {
	factory := NewEvmNode(config.EvmNode.Rpc, config.EvmNode.Wss, config.EvmNode.MaxScanHeight, config.EvmNode.ListenType)
	go dealFactory(factory)
	time.Sleep(time.Second)
}

func dealFactory(factory *EvmNode) {
	c, err := ethclient.Dial(config.EvmNode.Rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}
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
				addCoin(c, p.Token0)
				addCoin(c, p.Token1)
				initPoolStatReal(p.Contract)
				StartPool([]common.Address{common.HexToAddress(p.Contract)}, []common.Address{common.HexToAddress(p.Token0)}, []common.Address{common.HexToAddress(p.Token1)})
			}
		}
	}
}

func addCoin(client *ethclient.Client, contract string) {
	coins := model.GetCoins()
	for _, coin := range coins {
		if coin.Contract == contract {
			return
		}
	}
	erc20, err := contracts.NewERC20(common.HexToAddress(contract), client)
	if err != nil {
		gl.OutLogger.Error("NewERC20 error. %v : %v", contract, err)
		return
	}
	deci, err := erc20.Decimals(nil)
	if err != nil {
		gl.OutLogger.Error("erc20.Decimals(nil) error. %v : %v", contract, err)
		return
	}
	symbol, err := erc20.Symbol(nil)
	if err != nil {
		gl.OutLogger.Error("erc20.Symbol(nil) error. %v : %v", contract, err)
		return
	}
	if err := model.AddToken(symbol, contract, symbol, int64(deci), 1, 0); err != nil {
		gl.OutLogger.Error("Add coin to db error. %v : %v", contract, err)
		return
	}
	gl.OutLogger.Info("Add a new coin to db. %s : %s : %d", symbol, contract, deci)
}
