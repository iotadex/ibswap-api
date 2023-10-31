package api

import (
	"ibdex/gl"
	"ibdex/model"
	"ibdex/service"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func AddPool(c *gin.Context) {
	contract := c.PostForm("contract")
	version, _ := strconv.ParseInt(c.PostForm("version"), 10, 64)
	token0 := c.PostForm("token0")
	token1 := c.PostForm("token1")
	feeRate, _ := strconv.ParseInt(c.PostForm("fee_rate"), 10, 64)
	if len(contract) == 0 || version == 0 || len(token0) != 42 || len(token1) != 42 || feeRate == 0 {
		gl.OutLogger.Error("Add pool params error. %s : %d : %s : %s : %d", contract, version, token0, token1, feeRate)
		c.String(http.StatusOK, "params error")
		return
	}
	if p, err := model.AddPool(contract, version, token0, token1, int(feeRate)); err != nil {
		gl.OutLogger.Error("Add token to db error. %v", err)
		c.String(http.StatusOK, "params error")
	} else {
		c.String(http.StatusOK, "OK")
		service.StartPool([]common.Address{common.HexToAddress(p.Contract)}, []common.Address{common.HexToAddress(p.Token0)}, []common.Address{common.HexToAddress(p.Token1)})

	}
}

func ChangePoolState(c *gin.Context) {
	contract := c.PostForm("contract")
	state, err := strconv.Atoi(c.PostForm("state"))
	if len(contract) == 0 || err != nil {
		gl.OutLogger.Error("ChangePoolState params error. %s : %s : %v", contract, c.PostForm("state"), err)
		c.String(http.StatusOK, "params error")
		return
	}

	if err := model.ChangePoolState(contract, state); err != nil {
		gl.OutLogger.Error("ChangePoolState in db error. %s:%d:%v", contract, state, err)
		c.String(http.StatusOK, "pool maybe not exist")
	} else {
		c.String(http.StatusOK, "OK")
		gl.OutLogger.Info("ChangePoolState is successful. %s : %d", contract, state)
	}
}

func ChangeTokenPublic(c *gin.Context) {
	contract := c.PostForm("contract")
	public, err := strconv.Atoi(c.PostForm("public"))
	if len(contract) == 0 || err != nil {
		gl.OutLogger.Error("ChangeTokenPublic params error. %s : %s : %v", contract, c.PostForm("public"), err)
		c.String(http.StatusOK, "params error")
		return
	}

	if err := model.ChangeTokenPublic(contract, public); err != nil {
		gl.OutLogger.Error("ChangeTokenPublic in db error. %s:%d:%v", contract, public, err)
		c.String(http.StatusOK, "token maybe not exist")
	} else {
		c.String(http.StatusOK, "OK")
		gl.OutLogger.Info("ChangeTokenPublic is successful. %s : %d", contract, public)
	}
}
