package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

var (
	ctx       = context.Background()
	rdbClient *redis.Client
)

// redis connection examples from: https://tutorialedge.net/golang/go-redis-tutorial/
// pub-sub modified from: https://www.golangdev.in/2021/08/redis-publish-subscribe-example-using.html

func initRedisConnection() *redis.Client {

	rdbClient := redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6379",
		Password: "",
		DB:       0, // redis' default database name
	})

	pong, err := rdbClient.Ping().Result() // check if redis db is reachable
	fmt.Println(pong, err)

	return rdbClient

}

//publishArchiveTask marshals a struct into json
// and publishes to a redis channel where the task will be
// handled by a subscriber/worker
func publishArchiveTask(archiveTask archiveRequest) error {

	//convert struct to json
	jsonMessage, err := json.Marshal(archiveTask)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
	}
	//publish to channel
	channelName := "archives"
	err = rdbClient.Publish(channelName, jsonMessage).Err()
	if nil != err {
		fmt.Printf("Publish Error: %s", err.Error())
		return err
	}

	fmt.Println("published archive task")
	return nil

}
