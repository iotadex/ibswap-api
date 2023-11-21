package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"ibdex/config"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

var zeroAddress common.Address

type PoolStat struct {
	Address  string
	Tx       string
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

// EvmNode
type EvmNode struct {
	rpc           string
	wss           string
	maxScanHeight uint64
	listenType    int
}

func NewEvmNode(rpc, wss string, maxScanHeight uint64, lT int) *EvmNode {
	return &EvmNode{
		rpc:           rpc,
		wss:           wss,
		maxScanHeight: maxScanHeight,
		listenType:    lT,
	}
}

func (t *EvmNode) StartListenV2(pools []common.Address) (chan string, chan PoolStat) {
	//Set the query filter
	query := ethereum.FilterQuery{
		Addresses: pools,
		Topics:    [][]common.Hash{{EventSwapV2, EventSyncV2}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chPoolStat := make(chan PoolStat, 10)

	if t.listenType == 0 {
		go t.listenPoolsV2(query, chLog, chPoolStat)
	} else {
		go t.scanPoolsV2(query, chLog, chPoolStat)
	}

	return chLog, chPoolStat
}

func (t *EvmNode) listenPoolsV2(query ethereum.FilterQuery, chLog chan string, chPoolStat chan PoolStat) {
StartFilter:
	//Create the ethclient
	c, err := ethclient.Dial(t.wss)
	if err != nil {
		chLog <- fmt.Sprintf("The EthWssClient redial error(%v). \nThe EthWssClient will be redialed at 5 seconds later...", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	eventLogChan := make(chan types.Log)
	sub, err := c.SubscribeFilterLogs(context.Background(), query, eventLogChan)
	if err != nil || sub == nil {
		chLog <- fmt.Sprintf("Get event logs from eth wss client error. %v", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	log.Default().Printf("Start to listen V2 pool %d ...\n", len(query.Addresses))
	for {
		select {
		case err := <-sub.Err():
			chLog <- fmt.Sprintf("Event wss sub error(%v). \nThe EthWssClient will be redialed ...", err)
			sub.Unsubscribe()
			time.Sleep(time.Second * 5)
			goto StartFilter
		case vLog := <-eventLogChan:
			t.dealPoolV2Log(&vLog, chLog, chPoolStat)
		}
	}
}

func (t *EvmNode) scanPoolsV2(query ethereum.FilterQuery, chLog chan string, chPoolStat chan PoolStat) {
	c, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}
	fromHeight, err := c.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	log.Default().Printf("Start to scan V2 %d ...\n", len(query.Addresses))
	for {
		time.Sleep(config.ScanTime * time.Second)
		var toHeight uint64
		if toHeight, err = c.BlockNumber(context.Background()); err != nil {
			str := cutErrorString(err.Error())
			chLog <- fmt.Sprintf("BlockNumber error. %v", str)
			continue
		} else if toHeight < fromHeight {
			continue
		} else if toHeight-fromHeight > 10000 {
			toHeight = fromHeight + 9999
		}

		query.FromBlock = new(big.Int).SetUint64(fromHeight)
		query.ToBlock = new(big.Int).SetUint64(toHeight)
		logs, err := c.FilterLogs(context.Background(), query)
		if err != nil {
			chLog <- fmt.Sprintf("FilterLogs error. %v", err)
			continue
		}
		for i := range logs {
			t.dealPoolV2Log(&logs[i], chLog, chPoolStat)
		}
		fromHeight = toHeight + 1
	}
}

func (t *EvmNode) dealPoolV2Log(vLog *types.Log, chLog chan string, chPoolStat chan PoolStat) {
	amount0 := big.NewInt(0)
	amount1 := big.NewInt(0)
	reserve0 := big.NewInt(0)
	reserve1 := big.NewInt(0)
	if bytes.Equal(vLog.Topics[0].Bytes(), EventSyncV2.Bytes()) {
		reserve0 = new(big.Int).SetBytes(vLog.Data[:32])
		reserve1 = new(big.Int).SetBytes(vLog.Data[32:64])
	} else if bytes.Equal(vLog.Topics[0].Bytes(), EventSwapV2.Bytes()) {
		amount0 = new(big.Int).SetBytes(vLog.Data[:32])
		amount1 = new(big.Int).SetBytes(vLog.Data[32:64])
		amount0 = amount0.Add(amount0, new(big.Int).SetBytes(vLog.Data[64:96]))
		amount1 = amount1.Add(amount1, new(big.Int).SetBytes(vLog.Data[96:128]))
	}
	chPoolStat <- PoolStat{
		Reserve0: reserve0,
		Reserve1: reserve1,
		Volume0:  amount0,
		Volume1:  amount1,
		Tick:     0,
		Ts:       time.Now().Unix(),
	}
}

func (t *EvmNode) StartListenV3(pools []common.Address, token0s []common.Address, token1s []common.Address) (chan string, chan PoolStat) {
	c, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}

	erc20Tokens := make(map[string][2]*ERC20)
	for i, pool := range pools {
		token0, err := NewERC20(token0s[i], c)
		if err != nil {
			panic(err)
		}
		token1, err := NewERC20(token1s[i], c)
		if err != nil {
			panic(err)
		}
		erc20Tokens[pool.Hex()] = [2]*ERC20{token0, token1}
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chPoolStat := make(chan PoolStat, 10)

	//Set the query filter
	query := ethereum.FilterQuery{
		Addresses: pools,
		Topics:    [][]common.Hash{{EventSwapV3, EventMintV3, EventBurnV3}},
	}

	if t.listenType == 0 {
		go t.listenPoolsV3(erc20Tokens, query, chLog, chPoolStat)
	} else {
		go t.scanPoolsV3(erc20Tokens, query, chLog, chPoolStat)
	}

	return chLog, chPoolStat
}

func (t *EvmNode) listenPoolsV3(erc20Tokens map[string][2]*ERC20, query ethereum.FilterQuery, chLog chan string, chPoolStat chan PoolStat) {
StartFilter:
	//Create the ethclient
	c, err := ethclient.Dial(t.wss)
	if err != nil {
		chLog <- fmt.Sprintf("The EthWssClient redial error(%v). \nThe EthWssClient will be redialed at 5 seconds later...", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	eventLogChan := make(chan types.Log)
	sub, err := c.SubscribeFilterLogs(context.Background(), query, eventLogChan)
	if err != nil || sub == nil {
		chLog <- fmt.Sprintf("Get event logs from eth wss client error. %v", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	log.Default().Printf("Start to listen V3 pool %d ...\n", len(query.Addresses))
	for {
		select {
		case err := <-sub.Err():
			chLog <- fmt.Sprintf("Event wss sub error(%v). \nThe EthWssClient will be redialed ...", err)
			sub.Unsubscribe()
			time.Sleep(time.Second * 5)
			goto StartFilter
		case vLog := <-eventLogChan:
			tokens := erc20Tokens[vLog.Address.Hex()]
			t.dealPoolV3Log(tokens[0], tokens[1], &vLog, chLog, chPoolStat)
		}
	}
}

func (t *EvmNode) scanPoolsV3(erc20Tokens map[string][2]*ERC20, query ethereum.FilterQuery, chLog chan string, chPoolStat chan PoolStat) {
	c, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}

	fromHeight, err := c.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	log.Default().Printf("Start to scan V3 pool %d ...\n", len(query.Addresses))
	for {
		time.Sleep(config.ScanTime * time.Second)
		var toHeight uint64
		if toHeight, err = c.BlockNumber(context.Background()); err != nil {
			str := cutErrorString(err.Error())
			chLog <- fmt.Sprintf("BlockNumber error. %v", str)
			continue
		} else if toHeight < fromHeight {
			continue
		} else if toHeight-fromHeight > t.maxScanHeight {
			toHeight = fromHeight + t.maxScanHeight
		}

		query.FromBlock = new(big.Int).SetUint64(fromHeight)
		query.ToBlock = new(big.Int).SetUint64(toHeight)
		logs, err := c.FilterLogs(context.Background(), query)
		if err != nil {
			chLog <- fmt.Sprintf("FilterLogs error. %v", err)
			continue
		}
		for i := range logs {
			tokens := erc20Tokens[logs[i].Address.Hex()]
			t.dealPoolV3Log(tokens[0], tokens[1], &logs[i], chLog, chPoolStat)
		}
		fromHeight = toHeight + 1
	}
}

func (t *EvmNode) dealPoolV3Log(token0, token1 *ERC20, vLog *types.Log, chLog chan string, chPoolStat chan PoolStat) {
	amount0 := big.NewInt(0)
	amount1 := big.NewInt(0)
	var tick int64
	if bytes.Equal(vLog.Topics[0].Bytes(), EventSwapV3.Bytes()) {
		if len(vLog.Data) < 160 {
			chLog <- fmt.Sprintf("Swap Event Data error. %v", hex.EncodeToString(vLog.Data))
			return
		}
		amount0 = absBytes(vLog.Data[:32])
		amount1 = absBytes(vLog.Data[32:64])
		tick = flagBytes(vLog.Data[156:]).Int64()
	}
	reserve0 := big.NewInt(0)
	reserve1 := big.NewInt(0)
	if balance, err := token0.BalanceOf(&bind.CallOpts{}, vLog.Address); err != nil {
		chLog <- fmt.Sprintf("Balance of token0 error. %v", err)
	} else {
		reserve0 = balance
	}
	if balance, err := token1.BalanceOf(&bind.CallOpts{}, vLog.Address); err != nil {
		chLog <- fmt.Sprintf("Balance of token1 error. %v", err)
	} else {
		reserve1 = balance
	}
	chPoolStat <- PoolStat{
		Address:  vLog.Address.Hex(),
		Tx:       vLog.TxHash.Hex(),
		Reserve0: reserve0,
		Reserve1: reserve1,
		Volume0:  amount0,
		Volume1:  amount1,
		Tick:     tick,
		Ts:       time.Now().Unix(),
	}
}

func (t *EvmNode) StartListenNft(nftPos, f, code string) (chan string, chan NftToken) {
	factory := common.HexToAddress(f)
	initCode := common.FromHex(code)
	rcpClient, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}

	nft := common.HexToAddress(nftPos)

	iNftPostion, err := NewINonfungiblePositionManager(nft, rcpClient)
	if err != nil {
		panic(err)
	}

	//Set the query filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{nft},
		Topics:    [][]common.Hash{{EventTransferNFT}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chNft := make(chan NftToken, 10)

	if t.listenType == 0 {
		go t.listenNft(factory, initCode, iNftPostion, query, chLog, chNft)
	} else {
		go t.scanNft(factory, initCode, iNftPostion, query, chLog, chNft)
	}

	return chLog, chNft
}

func (t *EvmNode) listenNft(factory common.Address, initCode []byte, iNftPostion *INonfungiblePositionManager, query ethereum.FilterQuery, chLog chan string, chNft chan NftToken) {
StartFilter:
	//Create the ethclient
	c, err := ethclient.Dial(t.wss)
	if err != nil {
		chLog <- fmt.Sprintf("The EthWssClient redial error(%v). \nThe EthWssClient will be redialed at 5 seconds later...", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	eventLogChan := make(chan types.Log)
	sub, err := c.SubscribeFilterLogs(context.Background(), query, eventLogChan)
	if err != nil || sub == nil {
		chLog <- fmt.Sprintf("Get event logs from eth wss client error. %v", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	log.Default().Printf("Start to listen NFT pool %s ...\n", query.Addresses[0].Hex())
	for {
		select {
		case err := <-sub.Err():
			chLog <- fmt.Sprintf("Event wss sub error(%s:%v). \nThe EthWssClient will be redialed ...", factory.Hex(), err)
			sub.Unsubscribe()
			time.Sleep(time.Second * 5)
			goto StartFilter
		case vLog := <-eventLogChan:
			t.dealNFTLog(factory, initCode, iNftPostion, &vLog, chLog, chNft)
		}
	}
}

func (t *EvmNode) scanNft(factory common.Address, initCode []byte, iNftPostion *INonfungiblePositionManager, query ethereum.FilterQuery, chLog chan string, chNft chan NftToken) (chan string, chan NftToken) {
	c, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}

	fromHeight, err := c.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	fromHeight -= 7000

	log.Default().Printf("Start to scan NFT %s ...\n", query.Addresses[0].Hex())
	for {
		time.Sleep(config.ScanTime * time.Second)
		var toHeight uint64
		if toHeight, err = c.BlockNumber(context.Background()); err != nil {
			str := cutErrorString(err.Error())
			chLog <- fmt.Sprintf("BlockNumber error. %v", str)
			continue
		} else if toHeight < fromHeight {
			continue
		} else if toHeight-fromHeight > t.maxScanHeight {
			toHeight = fromHeight + t.maxScanHeight
		}

		query.FromBlock = new(big.Int).SetUint64(fromHeight)
		query.ToBlock = new(big.Int).SetUint64(toHeight)
		logs, err := c.FilterLogs(context.Background(), query)
		if err != nil {
			chLog <- fmt.Sprintf("FilterLogs error. %v", err)
			continue
		}
		for i := range logs {
			t.dealNFTLog(factory, initCode, iNftPostion, &logs[i], chLog, chNft)
		}
		fromHeight = toHeight + 1
	}
}

func (t *EvmNode) dealNFTLog(factory common.Address, initCode []byte, iNftPostion *INonfungiblePositionManager, vLog *types.Log, chLog chan string, chNft chan NftToken) {
	from := common.BytesToAddress(vLog.Topics[1].Bytes())
	to := common.BytesToAddress(vLog.Topics[2].Bytes())
	tokenId := new(big.Int).SetBytes(vLog.Topics[3].Bytes())
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
	if d == 0 {
		//neither Mint nor Burn
		return
	}
	//Get the postion
	var token0, token1, pool string
	var fee int
	if p, err := iNftPostion.Positions(&bind.CallOpts{}, tokenId); err != nil {
		chLog <- fmt.Sprintf("call positions from NewINonfungiblePositionManager error. %s : %v", tokenId.String(), err)
		return
	} else {
		token0, token1 = p.Token0.Hex(), p.Token1.Hex()
		fee = int(p.Fee.Int64())
		pool = getPoolContract(p.Token0, p.Token1, factory, p.Fee, initCode)
	}
	//Get the pool's contract address
	chNft <- NftToken{
		collection: vLog.Address.Hex(),
		pool:       pool,
		token0:     token0,
		token1:     token1,
		fee:        fee,
		user:       user,
		tokenId:    tokenId.String(),
		direction:  d,
	}
}

func (t *EvmNode) StartListenFactory(factory string) (chan string, chan Pool) {
	//Set the query filter
	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(factory)},
		Topics:    [][]common.Hash{{EventPoolCreated}},
	}

	// connetion err chan
	chLog := make(chan string, 10)
	chPool := make(chan Pool, 10)

	if t.listenType == 0 {
		go t.listenFactory(query, chLog, chPool)
	} else {
		go t.scanFactory(query, chLog, chPool)
	}
	return chLog, chPool
}

func (t *EvmNode) listenFactory(query ethereum.FilterQuery, chLog chan string, chPool chan Pool) {
StartFilter:
	//Create the ethclient
	c, err := ethclient.Dial(t.wss)
	if err != nil {
		chLog <- fmt.Sprintf("The EthWssClient redial error(%v). \nThe EthWssClient will be redialed at 5 seconds later...", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	eventLogChan := make(chan types.Log)
	sub, err := c.SubscribeFilterLogs(context.Background(), query, eventLogChan)
	if err != nil || sub == nil {
		chLog <- fmt.Sprintf("Get event logs from eth wss client error. %v", err)
		time.Sleep(time.Second * 5)
		goto StartFilter
	}
	log.Default().Printf("Start to listen Factory pool %s ...\n", query.Addresses[0].Hex())
	for {
		select {
		case err := <-sub.Err():
			chLog <- fmt.Sprintf("Event wss sub error(%v). \nThe EthWssClient will be redialed ...", err)
			sub.Unsubscribe()
			time.Sleep(time.Second * 5)
			goto StartFilter
		case vLog := <-eventLogChan:
			t.dealFactoryLog(&vLog, chPool)
		}
	}
}

func (t *EvmNode) scanFactory(query ethereum.FilterQuery, chLog chan string, chPool chan Pool) {
	c, err := ethclient.Dial(t.rpc)
	if err != nil {
		panic(err)
	}

	fromHeight, err := c.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}

	log.Default().Printf("Start to scan Factory %s ...\n", query.Addresses[0].Hex())
	for {
		time.Sleep(config.ScanTime * time.Second)
		var toHeight uint64
		if toHeight, err = c.BlockNumber(context.Background()); err != nil {
			str := cutErrorString(err.Error())
			chLog <- fmt.Sprintf("BlockNumber error. %v", str)
			continue
		} else if toHeight < fromHeight {
			continue
		} else if toHeight-fromHeight > t.maxScanHeight {
			toHeight = fromHeight + t.maxScanHeight
		}

		query.FromBlock = new(big.Int).SetUint64(fromHeight)
		query.ToBlock = new(big.Int).SetUint64(toHeight)
		logs, err := c.FilterLogs(context.Background(), query)
		if err != nil {
			chLog <- fmt.Sprintf("FilterLogs error. %v", err)
			continue
		}
		for i := range logs {
			t.dealFactoryLog(&logs[i], chPool)
		}
		fromHeight = toHeight + 1
	}
}

func (t *EvmNode) dealFactoryLog(vLog *types.Log, chPool chan Pool) {
	token0 := common.BytesToAddress(vLog.Topics[1].Bytes())
	token1 := common.BytesToAddress(vLog.Topics[2].Bytes())
	fee := new(big.Int).SetBytes(vLog.Topics[3].Bytes()).Int64()
	tickSpacing := new(big.Int).SetBytes(vLog.Data[:32]).Int64()
	pool := common.BytesToAddress(vLog.Data[32:])
	chPool <- Pool{
		Token0:      token0.Hex(),
		Token1:      token1.Hex(),
		FeeRate:     int(fee),
		TickSpacing: tickSpacing,
		Contract:    pool.Hex(),
	}
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

func cutErrorString(str string) string {
	bi := strings.Index(str, "<title>") + 7
	ei := strings.Index(str, "</title>")
	if len(str) > bi && len(str) > ei && bi < ei {
		str = str[bi:ei]
	}
	return str
}
