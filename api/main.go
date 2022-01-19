package main

import (
	conf "github.com/eth-library-lab/dataset-dj/configuration"
	"github.com/eth-library-lab/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
)

var (
	config *conf.ServerConfig
	runfig *conf.RuntimeConfig
)

func main() {

	config = conf.InitServerConfig()
	runfig = conf.InitRuntimeConfig(config)

	// Release db resource when the main
	// function is returned.
	defer dbutil.CloseMDB(runfig.MongoClient, runfig.MongoCtx, runfig.CtxCancel)

	router := gin.Default()
	router.GET("/files", getAvailableFiles)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.GET("/check", healthCheck)
	router.POST("/addSourceBucket", addSourceBucket)
	router.GET("/key", handleTokenRequest)
	router.POST("/key", handleValidateAPIToken)
	router.Run("0.0.0.0:" + config.Port) // bind to 0.0.0.0 to receive requests from outside docker container

}
