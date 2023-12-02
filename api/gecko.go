package api

import (
	"ibdex/gl"
	"ibdex/model"
	"ibdex/service"
	"math"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Pair struct {
	TickerId string `json:"ticker_id"`
	Base     string `json:"base"`
	Target   string `json:"target"`
	PoolId   string `json:"pool_id"`
}

func Pairs(c *gin.Context) {
	pools := model.GetPools(3)
	pairs := make([]Pair, 0)
	for _, p := range pools {
		base, err := model.GetCoin(p.Token0)
		if err != nil {
			gl.OutLogger.Error("GetCoin error. %s", p.Token0)
			continue
		}
		target, err := model.GetCoin(p.Token1)
		if err != nil {
			gl.OutLogger.Error("GetCoin error. %s", p.Token1)
			continue
		}
		pair := Pair{
			TickerId: base.Symbol + "_" + target.Symbol,
			Base:     base.Symbol + "/" + base.Contract,
			Target:   target.Symbol + "/" + target.Contract,
			PoolId:   p.Contract,
		}
		pairs = append(pairs, pair)
	}
	c.JSON(http.StatusOK, pairs)
}

type Ticker struct {
	TickerId       string `json:"ticker_id"`
	BaseCurrency   string `json:"base_currency"`
	TargetCurrency string `json:"target_currency"`
	PoolId         string `json:"pool_id"`
	LastPrice      string `json:"last_price"`
	BaseVolume     string `json:"base_volume"`
	TargetVolume   string `json:"target_volume"`
	LiquidityInUsd string `json:"liquidity_in_usd"`
	Bid            string `json:"bid"`
	Ask            string `json:"ask"`
	High           string `json:"high"`
	Low            string `json:"low"`
}

func Tickers(c *gin.Context) {
	pools := model.GetPools(3)
	tickers := make([]Ticker, 0)
	for _, p := range pools {
		base, err := model.GetCoin(p.Token0)
		if err != nil {
			gl.OutLogger.Error("GetCoin error. %s", p.Token0)
			continue
		}
		target, err := model.GetCoin(p.Token1)
		if err != nil {
			gl.OutLogger.Error("GetCoin error. %s", p.Token1)
			continue
		}

		overView, err := service.OverviewPoolsByContract(p.Contract)
		if err != nil {
			gl.OutLogger.Error("OverviewPoolsByContract error. %s", p.Contract)
			continue
		}

		v0, _ := new(big.Float).SetString(overView.Volume24H0)
		v, _ := v0.Float64()
		volume0 := v / math.Pow10(base.Decimal)
		v1, _ := new(big.Float).SetString(overView.Volume24H1)
		v, _ = v1.Float64()
		volume1 := v / math.Pow10(target.Decimal)

		r0, _ := new(big.Int).SetString(overView.Reserve0, 10)
		r1, _ := new(big.Int).SetString(overView.Reserve1, 10)
		r0.Mul(r0, r1)
		r0.Div(r0, big.NewInt(int64(math.Pow10(base.Decimal))))
		r0.Div(r0, big.NewInt(int64(math.Pow10(target.Decimal))))
		l := uint64(math.Sqrt(float64(r0.Uint64()) * service.GetTokenUsdPrice(p.Token0) * service.GetTokenUsdPrice(p.Token1)))

		ticker := Ticker{
			TickerId:       p.Contract,
			BaseCurrency:   base.Contract,
			TargetCurrency: target.Contract,
			PoolId:         p.Contract,
			LastPrice:      strconv.FormatFloat(math.Pow(1.0001, float64(overView.CurrTick))/math.Pow10(base.Decimal)*math.Pow10(target.Decimal), 'f', -1, 64),
			BaseVolume:     strconv.FormatFloat(volume0, 'f', -1, 64),
			TargetVolume:   strconv.FormatFloat(volume1, 'f', -1, 64),
			LiquidityInUsd: strconv.FormatUint(l, 10),
		}
		tickers = append(tickers, ticker)
	}
	c.JSON(http.StatusOK, tickers)
}

type TokenUsdPrice struct {
	Contract string  `json:"contract"`
	Symbol   string  `json:"symbol"`
	Price    float64 `json:"price"`
}

func GetUsdPrices(c *gin.Context) {
	tokens := model.GetCoins()
	prices := make([]TokenUsdPrice, 0)
	for _, t := range tokens {
		prices = append(prices, TokenUsdPrice{
			Contract: t.Contract,
			Symbol:   t.Symbol,
			Price:    service.GetTokenUsdPrice(t.Contract),
		})
	}
	c.JSON(http.StatusOK, prices)
}
