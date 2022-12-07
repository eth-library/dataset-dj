package main

import (
	conf "github.com/eth-library/dataset-dj/configuration"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
)

var (
	config  *conf.ServerConfig
	runtime *conf.RuntimeConfig
)

func setupConfig() {
	config = conf.InitServerConfig()
	runtime = conf.InitRuntimeConfig(config)
	_ = initAdminToken(runtime.MongoCtx, runtime.MongoClient)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/ping", healthCheck)
	router.GET("/key/validate", handleValidateAPIToken) // temporary, for debug purposes
	router.GET("/key/claim/:id", claimKey)              // use a link to create a token
	authorized := router.Group("/")
	authorized.Use(AuthMiddleware("service"))
	{
		authorized.GET("/files", getAvailableFiles)
		authorized.GET("/archive/:id", inspectArchive)
		authorized.POST("/archive", handleArchive)
		authorized.GET("/key/replace", replaceToken)
	}
	admin := router.Group("/admin") // for use by admins only
	admin.Use(AuthMiddleware("admin"))
	{
		admin.POST("/registerHandler", registerTaskHandler)
		admin.POST("/createKeyLink", handleCreateLink)
		admin.POST("/revokeKey", revokeToken)
		admin.POST("/addSourceBucket", addSourceBucket)
	}
	system := router.Group("/system") // for use by Task-handlers only
	system.Use(AuthMiddleware("system"))
	{
		system.GET("/requests", listOrders)
	}

	return router
}

func main() {
	setupConfig()
	router := setupRouter()
	defer dbutil.CloseMDB(runtime.MongoClient, runtime.MongoCtx, runtime.CtxCancel)
	router.Run("0.0.0.0:" + config.Port) // bind to 0.0.0.0 to receive requests from outside docker container
}
