package api

import (
	"context"
	"errors"
	"ibswap/api/middleware"
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
			log.Panicf("listen: %v\n", err)
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
	api := gin.New()
	api.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: GinLogger}), gin.Recovery())
	{
		api.GET("/pools/all", GetAllV2Pools)
		api.GET("/pools/:contract", GetPoolByContract)
		api.GET("/pools/overview", OverviewAllV2Pools)
		api.GET("/pools/:contract/overview", OverviewPoolByContract)
		api.GET("/pools/:contract/time-stats", StatPoolByContract)

		//		api.GET("/pools/:chain_id/all", GetAllV2Pools)
		//		api.GET("/pools/:chain_id/:contract", GetPoolByContract)
		//		api.GET("/pools/:chain_id/overview", OverviewAllV2Pools)
		//		api.GET("/pools/:chain_id/:contract/overview", OverviewPoolByContract)
		//		api.GET("/pools/:chain_id/:contract/time-stats", StatPoolByContract)
	}
	coins := api.Group("/coins")
	{
		coins.POST("/add", AddToken)

		coins.GET("/all", GetAllTokens)
		coins.GET("/:contract", GetTokenByContract)

		//		coins.GET("/:chain_id/all", GetAllTokens)
		//		coins.GET("/:chain_id/:contract", GetTokenByContract)
	}
	v2 := api.Group("/v2")
	{
		v2.GET("/pools/all", GetAllV2Pools)
		v2.GET("/pools/:contract", GetPoolByContract)
		v2.GET("/pools/overview", OverviewAllV2Pools)
		v2.GET("/pools/:contract/overview", OverviewPoolByContract)
		v2.GET("/pools/:contract/time-stats", StatPoolByContract)

		//		v2.GET("/pools/:chain_id/all", GetAllV2Pools)
		//		v2.GET("/pools/:chain_id/:contract", GetPoolByContract)
		//		v2.GET("/pools/:chain_id/overview", OverviewAllV2Pools)
		//		v2.GET("/pools/:chain_id/:contract/overview", OverviewPoolByContract)
		//		v2.GET("/pools/:chain_id/:contract/time-stats", StatPoolByContract)
	}
	v3 := api.Group("/v3")
	{
		v3.GET("/pools/all", GetAllV3Pools)
		v3.GET("/pools/:contract", GetPoolByContract)
		v3.GET("/pools/overview", OverviewAllV3Pools)
		v3.GET("/pools/:contract/overview", OverviewPoolByContract)
		v3.GET("/pools/:contract/time-stats", StatPoolByContract)

		//		v3.GET("/pools/:chain_id/all", GetAllV3Pools)
		//		v3.GET("/pools/:chain_id/:contract", GetPoolByContract)
		//		v3.GET("/pools/:chain_id/overview", OverviewAllV3Pools)
		//		v3.GET("/pools/:chain_id/:contract/overview", OverviewPoolByContract)
		//		v3.GET("/pools/:chain_id/:contract/time-stats", StatPoolByContract)

		v3.GET("/nfts/:user/:collection", GetNftTokensByUser)
	}

	admin := api.Group("/admin").Use(middleware.VerifySignature)
	{
		admin.POST("/pool/add", AddPool)
	}

	return api
}
