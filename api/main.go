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

func setupConfig() {
	config = conf.InitServerConfig()
	runfig = conf.InitRuntimeConfig(config)
	_ = initAdminToken(runfig.MongoCtx, runfig.MongoClient)
}

func setupRouter() *gin.Engine {

	router := gin.Default()
	router.GET("/ping", healthCheck)
	router.GET("/key/validate", handleValidateAPIToken) //temporary, for debug purposes
	router.GET("/key/claim/:id", claimKey)              //use a link to create a token
	authorized := router.Group("/")
	authorized.Use(AuthMiddleware("service"))
	{
		authorized.GET("/files", getAvailableFiles)
		authorized.GET("/archive/:id", inspectArchive)
		authorized.POST("/archive", handleArchive)
		authorized.GET("/key/replace", replaceToken)
	}
	admin := router.Group("/admin") //for use by admins only
	admin.Use(AuthMiddleware("admin"))
	{
		admin.POST("/createKeyLink", handleCreateLink)
		admin.POST("/revokeKey", revokeToken)
		admin.POST("/addSourceBucket", addSourceBucket)
	}

	return router
}

func main() {
	setupConfig()
	router := setupRouter()
	defer dbutil.CloseMDB(runfig.MongoClient, runfig.MongoCtx, runfig.CtxCancel)
	router.Run("0.0.0.0:" + config.Port) // bind to 0.0.0.0 to receive requests from outside docker container
}
