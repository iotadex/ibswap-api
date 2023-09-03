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
	contract := c.PostForm("contract")
	code := c.PostForm("code")
	decimal, err0 := strconv.ParseInt(c.PostForm("decimal"), 10, 64)
	t, err1 := strconv.ParseInt(c.PostForm("type"), 10, 64)
	public, err2 := strconv.ParseInt(c.PostForm("public"), 10, 64)
	if len(symbol) == 0 || len(contract) == 0 || len(code) == 0 || err0 != nil || err1 != nil || err2 != nil {
		gl.OutLogger.Error("Add token params error. %s :  %s : %s : %v : %v : %v", symbol, contract, code, err0, err1, err2)
		c.String(http.StatusOK, "params error")
		return
	}
	err := model.AddToken(symbol, contract, code, decimal, t, public)
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

func GetTokenByContract(c *gin.Context) {
	contract := c.Param("contract")
	coin, err := model.GetCoin(contract)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "token is not exist : " + " : " + contract,
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

func GetPoolByContract(c *gin.Context) {
	contract := c.Param("contract")
	if pool, err := model.GetPool(contract); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "Get Pool : " + " : " + contract,
		})
	} else {
		c.JSON(http.StatusOK, *pool)
	}
}

func OverviewAllV2Pools(c *gin.Context) {
	ps := service.OverviewPools(2)
	c.JSON(http.StatusOK, ps)
}

func OverviewAllV3Pools(c *gin.Context) {
	ps := service.OverviewPools(3)
	c.JSON(http.StatusOK, ps)
}

func OverviewPoolByContract(c *gin.Context) {
	contract := c.Param("contract")
	if p, err := service.OverviewPoolsByContract(contract); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "Overview Pool is not exist: " + " : " + contract,
		})
	} else {
		c.JSON(http.StatusOK, *p)
	}
}

func StatPoolByContract(c *gin.Context) {
	contract := c.Param("contract")
	if ps, err := service.StatPoolVolumes(contract); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "Statistic Pool is not exist: " + " : " + contract,
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
