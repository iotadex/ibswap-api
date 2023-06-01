package service

import (
	"bytes"
	"context"
	"fmt"
	"ibswap/config"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var EventSwapV2 = crypto.Keccak256Hash([]byte("Swap(address,uint256,uint256,uint256,uint256,address)"))
var EventSyncV2 = crypto.Keccak256Hash([]byte("Sync(uint112,uint112)"))

type PoolStat struct {
	Reserve0 *big.Int
	Reserve1 *big.Int
	Volume0  *big.Int
	Volume1  *big.Int
	Ts       int64
}

// EvmPool
type EvmPool struct {
	nodeUrl  string
	chainid  int64
	contract common.Address
}

func NewEvmPool(chainid int64, url, con string) *EvmPool {
	return &EvmPool{
		chainid:  chainid,
		nodeUrl:  url,
		contract: common.HexToAddress(con),
	}
}

func (t *EvmPool) StartListen() (chan string, chan PoolStat) {
	c, err := ethclient.Dial(t.nodeUrl)
	if err != nil {
		panic(err)
	}
	fromHeight, err := c.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	//Set the query filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{t.contract},
		Topics:    [][]common.Hash{{EventSwapV2, EventSyncV2}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chOrder := make(chan PoolStat, 10)

	go func() {
		log.Default().Printf("Start to scan %d : %s ...\n", t.chainid, t.contract.Hex())
		for {
			time.Sleep(config.ScanTime * time.Second)
			var toHeight uint64
			if toHeight, err = c.BlockNumber(context.Background()); err != nil {
				chLog <- fmt.Sprintf("BlockNumber error. %v", err)
				continue
			} else if toHeight < fromHeight {
				continue
			}

			query.FromBlock = new(big.Int).SetUint64(fromHeight)
			query.ToBlock = new(big.Int).SetUint64(toHeight)
			logs, err := c.FilterLogs(context.Background(), query)
			if err != nil {
				chLog <- fmt.Sprintf("FilterLogs error. %v", err)
				continue
			}
			amount0 := big.NewInt(0)
			amount1 := big.NewInt(0)
			reserve0 := big.NewInt(0)
			reserve1 := big.NewInt(0)
			for i := range logs {
				if bytes.Equal(logs[i].Topics[0].Bytes(), EventSyncV2.Bytes()) {
					reserve0 = new(big.Int).SetBytes(logs[i].Data[:32])
					reserve1 = new(big.Int).SetBytes(logs[i].Data[32:64])
				} else if bytes.Equal(logs[i].Topics[0].Bytes(), EventSwapV2.Bytes()) {
					amount0In := new(big.Int).SetBytes(logs[i].Data[:32])
					amount1In := new(big.Int).SetBytes(logs[i].Data[32:64])
					amount0Out := new(big.Int).SetBytes(logs[i].Data[64:96])
					amount1Out := new(big.Int).SetBytes(logs[i].Data[96:128])
					amount0.Add(amount0, amount0In)
					amount0.Add(amount0, amount0Out)
					amount1.Add(amount1, amount1In)
					amount1.Add(amount1, amount1Out)
				}
			}
			fromHeight = toHeight + 1
			if len(logs) > 0 {
				chOrder <- PoolStat{
					Reserve0: reserve0,
					Reserve1: reserve1,
					Volume0:  amount0,
					Volume1:  amount0,
					Ts:       time.Now().Unix(),
				}
			}
		}
	}()
	return chLog, chOrder
}
