package service

import (
	"encoding/json"
	"ibdex/gl"
	"ibdex/model"
	"ibdex/tools"
	"math"
	"strconv"
	"sync"
	"time"
)

type TokenPrice struct {
	price float64
	sync.RWMutex
}

func (p *TokenPrice) Get() float64 {
	p.RLock()
	defer p.RUnlock()
	return p.price
}

func (p *TokenPrice) Set(price float64) {
	p.Lock()
	defer p.Unlock()
	p.price = price
}

var currentEthPrice = &TokenPrice{}

func RealEthPrice() {
	tokenPricesAsUsd = make(map[string]*TokenPrice)
	f := func() {
		price, err := GetEthPriceFromCoinBash()
		if err != nil {
			gl.OutLogger.Error("GetEthPriceFromCoinBash, %v", err)
			return
		}
		currentEthPrice.Set(price)
		setTokensPriceAsUsd()
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

var tokenPricesAsUsd map[string]*TokenPrice

func setTokensPriceAsUsd() {
	pools := model.GetPools(3)

	q := make([]string, 3)
	q[0] = "0xa158A39d00C79019A01A6E86c56E96C461334Eb0"
	tokenPricesAsUsd[q[0]] = currentEthPrice
	q[1] = "0xa4f8C7C1018b9dD3be5835bF00f335D9910aF6Bd"
	tokenPricesAsUsd[q[1]] = &TokenPrice{
		price: 1,
	}
	q[2] = "0xeCE555d37C37D55a6341b80cF35ef3BC57401d1A"
	tokenPricesAsUsd[q[2]] = &TokenPrice{
		price: 1,
	}

	visited := make(map[string]bool)
	visited[q[0]] = true
	visited[q[1]] = true
	visited[q[2]] = true
	for len(q) > 0 {
		t0 := q[0]
		price := tokenPricesAsUsd[t0].Get()
		for _, p := range pools {
			token0, err0 := model.GetCoin(p.Token0)
			token1, err1 := model.GetCoin(p.Token1)
			currPool, err2 := OverviewPoolsByContract(p.Contract)
			if err0 != nil || err1 != nil || err2 != nil {
				gl.OutLogger.Error("Caculate usd price error. %v,%v,%v", err0, err1, err2)
				continue
			}
			f := math.Pow(1.0001, float64(currPool.CurrTick))
			if f == 0 || f == math.Inf(1) {
				continue
			}
			if p.Token0 == t0 {
				if !visited[p.Token1] {
					visited[p.Token1] = true
					q = append(q, p.Token1)

					//caculate the price
					newPrice := price / f * math.Pow10(token1.Decimal) / math.Pow10(token0.Decimal)
					if tokenPricesAsUsd[token1.Contract] == nil {
						tokenPricesAsUsd[token1.Contract] = &TokenPrice{price: newPrice}
					} else {
						tokenPricesAsUsd[token1.Contract].Set(newPrice)
					}
				}
			} else if p.Token1 == t0 {
				if !visited[p.Token0] {
					visited[p.Token0] = true
					q = append(q, p.Token0)

					//caculate the price
					newPrice := price * f * math.Pow10(token0.Decimal) / math.Pow10(token1.Decimal)
					if tokenPricesAsUsd[token0.Contract] == nil {
						tokenPricesAsUsd[token0.Contract] = &TokenPrice{price: newPrice}
					} else {
						tokenPricesAsUsd[token0.Contract].Set(newPrice)
					}
				}
			}
		}
		q = q[1:]
	}
}
