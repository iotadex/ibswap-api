package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"ibswap/config"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var EventSwapV2 = crypto.Keccak256Hash([]byte("Swap(address,uint256,uint256,uint256,uint256,address)"))
var EventSyncV2 = crypto.Keccak256Hash([]byte("Sync(uint112,uint112)"))

var EventSwapV3 = crypto.Keccak256Hash([]byte("Swap(address,address,int256,int256,uint160,uint128,int24)"))
var EventMintV3 = crypto.Keccak256Hash([]byte("Mint(address,address,int24,int24,uint128,uint256,uint256)"))
var EventBurnV3 = crypto.Keccak256Hash([]byte("Burn(address,int24,int24,uint128,uint256,uint256)"))

// event Transfer(address indexed from, address indexed to, uint256 indexed tokenId);
var EventTransferNFT = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

type PoolStat struct {
	Reserve0 *big.Int
	Reserve1 *big.Int
	Volume0  *big.Int
	Volume1  *big.Int
	Tick     int64
	Ts       int64
}

type NftToken struct {
	collection string
	tokenId    string
	user       string
	token0     string
	token1     string
	fee        int
	direction  int //1: mint, -1: burn
}

// EvmPool
type EvmPool struct {
	nodeUrl  string
	chainid  int64
	contract common.Address
	token0   common.Address
	token1   common.Address
}

func NewEvmPool(chainid int64, url, con string) *EvmPool {
	return &EvmPool{
		chainid:  chainid,
		nodeUrl:  url,
		contract: common.HexToAddress(con),
	}
}

func (t *EvmPool) StartListen(v int8) (chan string, chan PoolStat) {
	if v == 2 {
		return t.startListenV2()
	}
	return t.startListenV3()
}

func (t *EvmPool) startListenV2() (chan string, chan PoolStat) {
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
				str := err.Error()
				bi := strings.Index(str, "<title>") + 7
				ei := strings.Index(str, "</title>")
				if len(str) > bi && len(str) > ei {
					str = str[bi:ei]
				}
				chLog <- fmt.Sprintf("BlockNumber error. %v", str)
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
					Volume1:  amount1,
					Tick:     0,
					Ts:       time.Now().Unix(),
				}
			}
		}
	}()
	return chLog, chOrder
}

func (t *EvmPool) startListenV3() (chan string, chan PoolStat) {
	c, err := ethclient.Dial(t.nodeUrl)
	if err != nil {
		panic(err)
	}
	token0, err := NewERC20(t.token0, c)
	if err != nil {
		panic(err)
	}
	token1, err := NewERC20(t.token1, c)
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
		Topics:    [][]common.Hash{{EventSwapV3, EventMintV3, EventBurnV3}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chOrder := make(chan PoolStat, 10)

	go func() {
		log.Default().Printf("Start to scan V3 pool %d : %s ...\n", t.chainid, t.contract.Hex())
		for {
			time.Sleep(config.ScanTime * time.Second)
			var toHeight uint64
			if toHeight, err = c.BlockNumber(context.Background()); err != nil {
				str := err.Error()
				bi := strings.Index(str, "<title>") + 7
				ei := strings.Index(str, "</title>")
				if len(str) > bi && len(str) > ei {
					str = str[bi:ei]
				}
				chLog <- fmt.Sprintf("BlockNumber error. %v", str)
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
			var tick int64
			for i := range logs {
				if bytes.Equal(logs[i].Topics[0].Bytes(), EventSwapV3.Bytes()) {
					if len(logs[i].Data) < 160 {
						chLog <- fmt.Sprintf("Swap Event Data error. %v", hex.EncodeToString(logs[i].Data))
					}
					amount0.Abs(new(big.Int).SetBytes(logs[i].Data[:32]))
					amount1.Abs(new(big.Int).SetBytes(logs[i].Data[32:64]))
					tick = new(big.Int).SetBytes(logs[i].Data[128:]).Int64()
				}
			}
			fromHeight = toHeight + 1
			if len(logs) > 0 {
				// Get balances from ERC20
				if balance, err := token0.BalanceOf(&bind.CallOpts{}, t.contract); err != nil {
					chLog <- fmt.Sprintf("Balance of token0 error. %v", err)
				} else {
					reserve0 = balance
				}
				if balance, err := token1.BalanceOf(&bind.CallOpts{}, t.contract); err != nil {
					chLog <- fmt.Sprintf("Balance of token1 error. %v", err)
				} else {
					reserve1 = balance
				}

				chOrder <- PoolStat{
					Reserve0: reserve0,
					Reserve1: reserve1,
					Volume0:  amount0,
					Volume1:  amount1,
					Tick:     tick,
					Ts:       time.Now().Unix(),
				}
			}
		}
	}()
	return chLog, chOrder
}

func (t *EvmPool) StartListenNFT() (chan string, chan NftToken) {
	c, err := ethclient.Dial(t.nodeUrl)
	if err != nil {
		panic(err)
	}
	iNftPostion, err := NewINonfungiblePositionManager(t.contract, c)
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
		Topics:    [][]common.Hash{{EventTransferNFT}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chNft := make(chan NftToken, 10)

	var zeroAddress common.Address

	go func() {
		log.Default().Printf("Start to scan NFT %d : %s ...\n", t.chainid, t.contract.Hex())
		for {
			time.Sleep(config.ScanTime * time.Second)
			var toHeight uint64
			if toHeight, err = c.BlockNumber(context.Background()); err != nil {
				str := err.Error()
				bi := strings.Index(str, "<title>") + 7
				ei := strings.Index(str, "</title>")
				if len(str) > bi && len(str) > ei {
					str = str[bi:ei]
				}
				chLog <- fmt.Sprintf("BlockNumber error. %v", str)
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
			for i := range logs {
				from := common.BytesToAddress(logs[i].Topics[1].Bytes())
				to := common.BytesToAddress(logs[i].Topics[2].Bytes())
				tokenId := new(big.Int).SetBytes(logs[i].Topics[3].Bytes())
				var d int
				var user string
				if bytes.Equal(from[:], zeroAddress[:]) { //Mint NFT
					d = 1
					user = to.String()
				}
				if bytes.Equal(to[:], zeroAddress[:]) { // Burn NFT
					d = -1
					user = from.String()
				}
				//Get the postion
				var token0, token1 string
				var fee int
				if p, err := iNftPostion.Positions(&bind.CallOpts{}, tokenId); err != nil {
					chLog <- fmt.Sprintf("call positions from NewINonfungiblePositionManager error. %s : %v", tokenId.String(), err)
				} else {
					token0, token1 = p.Token0.Hex(), p.Token1.Hex()
					fee = int(p.Fee.Int64())
				}
				chNft <- NftToken{
					collection: t.contract.String(),
					token0:     token0,
					token1:     token1,
					fee:        fee,
					user:       user,
					tokenId:    tokenId.String(),
					direction:  d,
				}
			}
			fromHeight = toHeight + 1
		}
	}()
	return chLog, chNft
}
