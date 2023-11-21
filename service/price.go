package service

import (
	"encoding/json"
	"ibdex/gl"
	"ibdex/tools"
	"strconv"
	"sync"
	"time"
)

type EthPrice struct {
	price float64
	sync.RWMutex
}

func (p *EthPrice) Get() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.price
}

func (p *EthPrice) Set(price float64) {
	p.Lock()
	defer p.Unlock()
	p.price = price
}

var currentEthPrice = &EthPrice{}

func RealEthPrice() {
	f := func() {
		price, err := GetEthPriceFromCoinBash()
		if err != nil {
			gl.OutLogger.Error("GetEthPriceFromCoinBash, %v", err)
			return
		}
		currentEthPrice.Set(price)
	}
	priceTicker := time.NewTicker(time.Second * 180)
	f()
	go func() {
		for range priceTicker.C {
			f()
		}
	}()
}

type CoinBasePrice struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

func GetEthPriceFromCoinBash() (float64, error) {
	data, err := tools.HttpRequest("https://api.coinbase.com/v2/prices/ETH-USD/spot", "GET", nil, nil)
	if err != nil {
		return 0, err
	}
	var cd CoinBasePrice
	if err := json.Unmarshal(data, &cd); err != nil {
		return 0, err
	}
	var price float64
	if price, err = strconv.ParseFloat(cd.Data.Amount, 64); err != nil {
		return 0, err
	}
	return price, nil
}
