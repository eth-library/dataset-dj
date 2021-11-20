package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis"
)

var (
	ctx       = context.Background()
	rdbClient *redis.Client
)

// redis connection from tutorial:
// https://tutorialedge.net/golang/go-redis-tutorial/

type metaArchive struct {
	ID          string   `json:"id"`
	Files       []string `json:"files"`
	TimeCreated string   `json:"timeCreated"`
	TimeUpdated string   `json:"timeUpdated"`
	Status      string   `json:"status"`
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
	channelName := "default"

	//can subscribe to messages from multiple channels
	subscriber := rdbClient.Subscribe(channelName)
	defer subscriber.Close()
	channel := subscriber.Channel()
	for msg := range channel {
		fmt.Println("received", msg.Payload, "from", msg.Channel)
	}
}

func main() {
	fmt.Println("started task subscriber")

	rdbClient = initRedisConnection()
	subscribeToRedisChannel(rdbClient)
}
