package api

import (
	"ibswap/gl"
	"ibswap/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

/*
type Pool struct {
	ChainID  int64  `json:"chainid"`
	Contract string `json:"contract"`
	Version  int8
	Token0   string `json:"token0"`
	Token1   string `json:"token1"`
	FeeRate  int    `json:"fee_rate"`
	Decimal  int    `json:"decimal"`
}
*/

func AddPool(c *gin.Context) {
	chainid, _ := strconv.ParseInt(c.PostForm("chainid"), 10, 64)
	contract := c.PostForm("contract")
	version, _ := strconv.ParseInt(c.PostForm("version"), 10, 64)
	token0 := c.PostForm("token0")
	token1 := c.PostForm("token1")
	feeRate, _ := strconv.ParseInt(c.PostForm("fee_rate"), 10, 64)
	if chainid == 0 || len(contract) == 0 || version == 0 || len(token0) != 42 || len(token1) != 42 || feeRate == 0 {
		gl.OutLogger.Error("Add pool params error. %d : %s : %d : %s : %s : %d", chainid, contract, version, token0, token1, feeRate)
		c.String(http.StatusOK, "params error")
		return
	}
	if err := model.AddPool(chainid, contract, version, token0, token1, feeRate); err != nil {
		gl.OutLogger.Error("Add token to db error. %v", err)
		c.String(http.StatusOK, "params error")
	} else {
		c.String(http.StatusOK, "OK")
	}
}
