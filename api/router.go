package api

import (
	"context"
	"errors"
	"ibswap/config"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/triplefi/go-logger/logger"
)

var httpServer *http.Server

func StartHttpServer() {
	router := InitRouter()
	httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(config.HttpPort),
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %v\n", err)
		}
	}()
}

func StopHttpServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// InitRouter init the router
func InitRouter() *gin.Engine {
	if err := os.MkdirAll("./logs/http", os.ModePerm); err != nil {
		log.Panicf("Create dir './logs/http' error. %v", err)
	}
	GinLogger, err := logger.New("logs/http/gin.log", 2, 100*1024*1024, 10)
	if err != nil {
		log.Panicf("Create GinLogger file error. %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	dex := gin.New()
	dex.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: GinLogger}), gin.Recovery())
	{
		dex.POST("/coins/add", AddToken)
		dex.GET("/coins/all", GetAllTokens)
		dex.GET("/coins/:chain_id/all", GetAllTokensByChain)
		dex.GET("/coins/:chain_id/:contract", GetTokenByChainAndContract)

		dex.GET("/pools/all", GetAllPools)
		dex.GET("/pools/:chain_id/all", GetAllPoolsByChain)
		dex.GET("/pools/:chain_id/:contract", GetPoolByChainAndContract)
		dex.GET("/pools/:chain_id/overview", OverviewAllPoolsByChain)
		dex.GET("/pools/:chain_id/:contract/overview", OverviewPoolByChainAndContract)
		dex.GET("/pools/:chain_id/:contract/time-stats", StatPoolByChainAndContract)
	}

	return dex
}

func Test(c *gin.Context) {
	c.String(http.StatusOK, "Test OK!")
}
