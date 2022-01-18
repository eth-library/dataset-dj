// this is package subscribes to the redis channel and asynchronously handles requests to zip files
package main

import (
	"fmt"

	conf "github.com/eth-library-lab/dataset-dj/configuration"
	"github.com/eth-library-lab/dataset-dj/redisutil"
)

var (
	config *conf.ServerConfig
	runfig *conf.RuntimeConfig
)

func main() {
	fmt.Println("started task subscriber")

	config = conf.InitServerConfig()
	runfig = conf.InitRuntimeConfig(config)

	handlers := map[string]interface{}{
		"archive":       handleArchiveMessage,
		"emails":        handleEmailMessage,
		"sourceBuckets": handleSourceBucketMessage,
	}

	redisutil.SubscribeToRedisChannel(runfig.RdbClient, handlers)

}
