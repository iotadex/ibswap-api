package api

import (
	"ibswap/gl"
	"ibswap/model"
	"ibswap/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AddToken(c *gin.Context) {
	symbol := c.PostForm("symbol")
	chainid, _ := strconv.ParseInt(c.PostForm("chainid"), 10, 64)
	contract := c.PostForm("contract")
	code := c.PostForm("code")
	decimal, err0 := strconv.ParseInt(c.PostForm("decimal"), 10, 64)
	t, err1 := strconv.ParseInt(c.PostForm("type"), 10, 64)
	public, err2 := strconv.ParseInt(c.PostForm("public"), 10, 64)
	if len(symbol) == 0 || chainid == 0 || len(contract) == 0 || len(code) == 0 || err0 != nil || err1 != nil || err2 != nil {
		gl.OutLogger.Error("Add token params error. %s : %d : %s : %s : %v : %v : %v", symbol, chainid, contract, code, err0, err1, err2)
		c.String(http.StatusOK, "params error")
		return
	}
	err := model.AddToken(symbol, chainid, contract, code, decimal, t, public)
	if err != nil {
		c.String(http.StatusOK, "OK")
	} else {
		gl.OutLogger.Error("Add token to db error. %v", err)
		c.String(http.StatusOK, "params error")
	}
}

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

func GetAllV2Pools(c *gin.Context) {
	c.JSON(http.StatusOK, model.GetPools(2))
}

func GetAllV3Pools(c *gin.Context) {
	c.JSON(http.StatusOK, model.GetPools(3))
}

func GetAllV2PoolsByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	c.JSON(http.StatusOK, model.GetPoolsByChainId(chainID, 2))
}

func GetAllV3PoolsByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	c.JSON(http.StatusOK, model.GetPoolsByChainId(chainID, 3))
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

func OverviewAllV2PoolsByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	ps := service.OverviewPoolsByChainid(chainID, 2)
	c.JSON(http.StatusOK, ps)
}

func OverviewAllV3PoolsByChain(c *gin.Context) {
	chainID, _ := strconv.ParseInt(c.Param("chain_id"), 10, 64)
	ps := service.OverviewPoolsByChainid(chainID, 3)
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

func GetNftTokensByUser(c *gin.Context) {
	user := c.Param("user")
	collection := c.Param("collection")
	if ts, err := model.GetNftTokens(user, collection); err != nil {
		gl.OutLogger.Error("GetNftTokens error. %s : %s : %v", user, collection, err)
		c.JSON(http.StatusOK, gin.H{
			"error": "There is no NftToken: " + user + " : " + collection,
		})
	} else {
		c.JSON(http.StatusOK, ts)
	}
}
