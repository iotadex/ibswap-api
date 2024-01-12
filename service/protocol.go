package service

import (
	"fmt"
	"ibdex/config"
	"ibdex/contracts"
	"ibdex/gl"
	"ibdex/model"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func RealProtocolFees() {
	rpcClient, err := ethclient.Dial(config.EvmNode.Rpc)
	if err != nil {
		panic(fmt.Errorf("dial node error. %v", err))
	}

	f := func(t time.Duration) {
		pools := model.GetPools(3)
		for _, pool := range pools {
			time.Sleep(time.Second * t)
			uniV3Pool, err := contracts.NewUniswapV3Pool(common.HexToAddress(pool.Contract), rpcClient)
			if err != nil {
				gl.OutLogger.Error("NewUniswapV3Pool error, %v", err)
				continue
			}
			fees, err := uniV3Pool.ProtocolFees(&bind.CallOpts{})
			if err != nil {
				gl.OutLogger.Error("uniV3Pool.ProtocolFees error, %v", err)
				continue
			}
			if err = model.StoreProtocolFees(pool.Contract, pool.Symbol0, fees.Token0.String(), pool.Symbol1, fees.Token1.String()); err != nil {
				gl.OutLogger.Error("model.StoreProtocolFees error, %v", err)
			}
		}
	}
	ticker := time.NewTicker(time.Hour)
	f(1)
	go func() {
		for range ticker.C {
			f(10)
		}
	}()
}
