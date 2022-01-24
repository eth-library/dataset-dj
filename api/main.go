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

func setupRouter() *gin.Engine {

	config = conf.InitServerConfig()
	runfig = conf.InitRuntimeConfig(config)
	_ = initAdminToken(runfig.MongoCtx, runfig.MongoClient)
	// Release db resource when the main
	// function is returned.

	router := gin.Default()
	router.GET("/check", healthCheck)
	router.GET("/key/validate", handleValidateAPIToken) //temporary, for debug purposes
	router.GET("key/claim/:id", claimKey)               //use a link to create a token
	authorized := router.Group("/")
	authorized.Use(AuthMiddleware("service"))
	{
		authorized.GET("/files", getAvailableFiles)
		authorized.GET("/archive/:id", inspectArchive)
		authorized.POST("/archive", handleArchive)
		authorized.POST("/addSourceBucket", addSourceBucket)
		authorized.GET("/key/replace", AuthMiddleware("service"), replaceToken)
	}
	admin := router.Group("/admin")
	admin.Use(AuthMiddleware("admin"))
	{
		admin.POST("/createKeyLink", handleCreateLink) //TO DO add Auth. for use by admins
	}

	return router
}

func main() {

	router := setupRouter()
	defer dbutil.CloseMDB(runfig.MongoClient, runfig.MongoCtx, runfig.CtxCancel)

	router.Run("0.0.0.0:" + config.Port) // bind to 0.0.0.0 to receive requests from outside docker container

}
