package main

import (
	conf "github.com/eth-library/dataset-dj/configuration"
	"github.com/eth-library/dataset-dj/dbutil"
	"github.com/gin-gonic/gin"
)

var (
	config  *conf.ApiConfig
	runtime *conf.ApiRuntime
)

func setupConfig() {
	config = conf.InitApiConfig()
	runtime = conf.InitApiRuntime(config)
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
		authorized.GET("/archive/:id", inspectArchive)
		authorized.POST("/archive", handleArchive)
		authorized.GET("/key/replace", replaceToken)
		authorized.POST("/source", registerSource)
	}
	admin := router.Group("/admin") // for use by admins only
	admin.Use(AuthMiddleware("admin"))
	{
		admin.POST("/registerHandler", registerTaskHandler)
		admin.POST("/createKeyLink", handleCreateLink)
		admin.POST("/revokeKey", revokeToken)
	}
	handler := router.Group("/handler") // for use by Task-handlers only
	handler.Use(AuthMiddleware("handler"))
	{
		handler.POST("/orders", listOrders)
		handler.POST("/order/:id", updateStatus)
		handler.GET("/archive/:id", inspectArchive) // resolves to /handler/archive/:id
	}

	return router
}

func main() {
	setupConfig()
	router := setupRouter()
	defer dbutil.CloseMDB(runtime.MongoClient, runtime.MongoCtx, runtime.CtxCancel)
	router.Run("0.0.0.0:" + config.Port) // bind to 0.0.0.0 to receive requests from outside docker container
}
