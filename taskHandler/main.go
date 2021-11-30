// this is package subscribes to the redis channel and asynchronously handles requests to zip files
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/go-redis/redis"
)

var (
	config        *serverConfig
	storageClient *storage.Client
	ctx           context.Context
	rdbClient     *redis.Client
)

// redis connection from tutorial:
// https://tutorialedge.net/golang/go-redis-tutorial/

type archiveRequest struct {
	Email     string   `json:"email"`
	ArchiveID string   `json:"archiveID"`
	Files     []string `json:"files"`
	Source    string   `json:"source"`
}

func initRedisConnection() *redis.Client {

	rdbClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0, // redis' default database name
	})

	// check if redis db is reachable
	pong, err := rdbClient.Ping().Result()
	fmt.Println(pong, err)

	return rdbClient

}

func subscribeToRedisChannel(rdbClient *redis.Client) {

	//can subscribe to messages from multiple channels
	subscriber := rdbClient.Subscribe("archives", "emails")
	defer subscriber.Close()
	channel := subscriber.Channel()

	var archRequest archiveRequest

	for msg := range channel {
		fmt.Println("received", msg.Payload, "from", msg.Channel)
		if msg.Channel == "archives" {

			// convert json string into struct
			json.Unmarshal([]byte(msg.Payload), &archRequest)

			fmt.Println("handling archRequest: ", archRequest)
			err := zipFiles(archRequest)
			if err != nil {
				fmt.Println("err: ", err)
			}

		}
	}
}

func main() {
	fmt.Println("started task subscriber")

	config = initServerConfig()
	ctx = context.Background()

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer storageClient.Close()

	rdbClient = initRedisConnection()
	subscribeToRedisChannel(rdbClient)

}
