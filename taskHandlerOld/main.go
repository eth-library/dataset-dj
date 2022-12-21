// this is package subscribes to the redis channel and asynchronously handles requests to zip files
package main

import (
	"fmt"

	conf "github.com/eth-library/dataset-dj/configuration"
	"github.com/eth-library/dataset-dj/redisutil"
)

var (
	config *conf.ServerConfig
	runfig *conf.RuntimeConfig
)

var subscribers = map[string]interface{}{
	"archives":      handleArchiveMessage,
	"emails":        handleEmailMessage,
	"sourceBuckets": handleSourceBucketMessage,
}

func main() {
	fmt.Println("started task subscriber")

	config = conf.InitServerConfig()
	runfig = conf.InitApiRuntime(config)

	redisutil.SubscribeToRedisChannel(runfig.RdbClient, subscribers)

}
