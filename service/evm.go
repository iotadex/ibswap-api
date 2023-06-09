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
// event IncreaseLiquidity(uint256 indexed tokenId,uint128 liquidity,uint256 amount0,uint256 amount1);
// event DecreaseLiquidity(uint256 indexed tokenId,uint128 liquidity,uint256 amount0,uint256 amount1);
var EventTransferNFT = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
var EventIncreaseLiq = crypto.Keccak256Hash([]byte("IncreaseLiquidity(uint256,uint128,uint256,uint256)"))
var EventDecreaseLiq = crypto.Keccak256Hash([]byte("DecreaseLiquidity(uint256,uint128,uint256,uint256)"))

// event PoolCreated(address indexed token0,address indexed token1,uint24 indexed fee,int24 tickSpacing,address pool)
var EventPoolCreated = crypto.Keccak256Hash([]byte("PoolCreated(address,address,uint24,int24,address)"))

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
	pool       string
	token0     string
	token1     string
	fee        int
	direction  int //1: mint, -1: burn
}

type Pool struct {
	Contract    string
	Token0      string
	Token1      string
	FeeRate     int
	TickSpacing int64
}

// EvmPool
type EvmPool struct {
	nodeUrl  string
	chainid  int64
	contract common.Address
	token0   common.Address
	token1   common.Address
}

func NewEvmPool(chainid int64, url, con, t0, t1 string) *EvmPool {
	return &EvmPool{
		chainid:  chainid,
		nodeUrl:  url,
		contract: common.HexToAddress(con),
		token0:   common.HexToAddress(t0),
		token1:   common.HexToAddress(t1),
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
		log.Default().Printf("Start to scan V2 %d : %s ...\n", t.chainid, t.contract.Hex())
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
			var tick int64
			for i := range logs {
				if bytes.Equal(logs[i].Topics[0].Bytes(), EventSwapV3.Bytes()) {
					if len(logs[i].Data) < 160 {
						chLog <- fmt.Sprintf("Swap Event Data error. %v", hex.EncodeToString(logs[i].Data))
						continue
					}
					amount0.Add(amount0, absBytes(logs[i].Data[:32]))
					amount1.Add(amount1, absBytes(logs[i].Data[32:64]))
					tick = flagBytes(logs[i].Data[156:]).Int64()
				}
			}
			fromHeight = toHeight + 1
			if len(logs) > 0 {
				// Get balances from ERC20
				reserve0 := big.NewInt(0)
				reserve1 := big.NewInt(0)
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

func (t *EvmPool) StartListenNFT(f, code string) (chan string, chan NftToken) {
	factory := common.HexToAddress(f)
	initCode := common.FromHex(code)
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
				var token0, token1, pool string
				var fee int
				if p, err := iNftPostion.Positions(&bind.CallOpts{}, tokenId); err != nil {
					chLog <- fmt.Sprintf("call positions from NewINonfungiblePositionManager error. %s : %v", tokenId.String(), err)
					continue
				} else {
					token0, token1 = p.Token0.Hex(), p.Token1.Hex()
					fee = int(p.Fee.Int64())
					pool = getPoolContract(p.Token0, p.Token1, factory, p.Fee, initCode)
				}
				//Get the pool's contract address
				chNft <- NftToken{
					collection: t.contract.String(),
					pool:       pool,
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

func (t *EvmPool) StartListenFactory() (chan string, chan Pool) {
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
		Topics:    [][]common.Hash{{EventPoolCreated}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chPool := make(chan Pool, 10)

	go func() {
		log.Default().Printf("Start to scan Factory %d : %s ...\n", t.chainid, t.contract.Hex())
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
				token0 := common.BytesToAddress(logs[i].Topics[1].Bytes())
				token1 := common.BytesToAddress(logs[i].Topics[2].Bytes())
				fee := new(big.Int).SetBytes(logs[i].Topics[3].Bytes()).Int64()
				tickSpacing := new(big.Int).SetBytes(logs[i].Data[:32]).Int64()
				pool := common.BytesToAddress(logs[i].Data[32:])
				chPool <- Pool{
					Token0:      token0.Hex(),
					Token1:      token1.Hex(),
					FeeRate:     int(fee),
					TickSpacing: tickSpacing,
					Contract:    pool.Hex(),
				}
			}
			fromHeight = toHeight + 1
		}
	}()
	return chLog, chPool
}

func getPoolContract(t0, t1, factory common.Address, fee *big.Int, initCode []byte) string {
	data := common.LeftPadBytes(t0.Bytes(), 32)
	data = append(data, common.LeftPadBytes(t1.Bytes(), 32)...)
	data = append(data, common.LeftPadBytes(fee.Bytes(), 32)...)
	s1 := crypto.Keccak256(data)

	d := []byte{15 + 15<<4} //'0xff'
	d = append(d, factory.Bytes()...)
	d = append(d, s1...)
	d = append(d, initCode...)
	s2 := crypto.Keccak256(d)

	return common.BytesToAddress(s2[12:]).Hex()
}

func absBytes(d []byte) *big.Int {
	if len(d) == 0 {
		return nil
	}
	a := big.NewInt(0)
	if d[0]&0x80 > 0 {
		for i := range d {
			d[i] = ^d[i]
		}
		a = new(big.Int).SetBytes(d)
		a.Add(a, big.NewInt(1))
	}
	return a
}

func flagBytes(d []byte) *big.Int {
	if len(d) == 0 {
		return nil
	}
	var a *big.Int
	if d[0]&0x80 > 0 {
		for i := range d {
			d[i] = ^d[i]
		}
		a = new(big.Int).SetBytes(d)
		a.Add(a, big.NewInt(1))
		a = a.Neg(a)
	} else {
		a = new(big.Int).SetBytes(d)
	}
	return a
}
