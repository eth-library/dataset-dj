package main

import (
	conf "github.com/eth-library-lab/dataset-dj/configuration"
	"github.com/gin-gonic/gin"
)

var (
	config *conf.ServerConfig
	runfig *conf.RuntimeConfig
)

func main() {

	config = conf.InitServerConfig()
	runfig = conf.InitRuntimeConfig(config)

	router := gin.Default()
	router.GET("/files", getAvailableFiles)
	router.GET("/archive/:id", inspectArchive)
	router.POST("/archive", handleArchive)
	router.GET("/check", healthCheck)
	router.POST("/addSourceBucket", addSourceBucket)
	router.Run("0.0.0.0:" + config.Port) // bind to 0.0.0.0 to receive requests from outside docker container

}
