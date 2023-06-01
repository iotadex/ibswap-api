package api

import (
	"ibswap/model"
	"ibswap/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllTokens(c *gin.Context) {
	coins := model.GetCoins()
	c.JSON(http.StatusOK, coins)
}

func GetAllTokensByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	coins := model.GetCoinsByChainId(chainID)
	c.JSON(http.StatusOK, coins)
}

func GetTokenByChainAndContract(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	contract := c.Param("contract")
	coin, err := model.GetCoin(chainID, contract)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "token is not exist : " + c.Param("chain_id") + " : " + contract,
		})
	} else {
		c.JSON(http.StatusOK, *coin)
	}
}

func GetAllPools(c *gin.Context) {
	c.JSON(http.StatusOK, model.GetPools())
}

func GetAllPoolsByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	c.JSON(http.StatusOK, model.GetPoolsByChainId(chainID))
}

func GetPoolByChainAndContract(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	contract := c.Param("contract")
	if pool, err := model.GetPool(chainID, contract); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "Get Pool : " + c.Param("chain_id") + " : " + contract,
		})
	} else {
		c.JSON(http.StatusOK, *pool)
	}
}

func OverviewAllPoolsByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	ps := service.OverviewPoolsByChainid(chainID)
	c.JSON(http.StatusOK, ps)
}

func OverviewPoolByChainAndContract(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	contract := c.Param("contract")
	if p, err := service.OverviewPoolsByChainidAndContract(chainID, contract); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "Overview Pool is not exist: " + c.Param("chain_id") + " : " + contract,
		})
	} else {
		c.JSON(http.StatusOK, *p)
	}
}

func StatPoolByChainAndContract(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	contract := c.Param("contract")
	if ps, err := service.CountPoolVolumes(chainID, contract); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "Statistic Pool is not exist: " + c.Param("chain_id") + " : " + contract,
		})
	} else {
		c.JSON(http.StatusOK, ps)
	}
}
