package redisutil

import (
	"encoding/json"
	"fmt"

	"github.com/go-redis/redis"
)

//EmailParts required by the mailHandler to send an email
type EmailParts struct {
	To       string
	Subject  string
	BodyType string // e.g.: text/plain
	Body     string
}

// redis connection examples from: https://tutorialedge.net/golang/go-redis-tutorial/
// pub-sub modified from: https://www.golangdev.in/2021/08/redis-publish-subscribe-example-using.html

//PublishTask marshals a struct into json
// and publishes to the channelName redis channel where the task will be
// handled by a subscriber/worker
func PublishTask(client *redis.Client, TaskMessage interface{}, channelName string) error {

	//convert struct to json
	jsonMessage, err := json.Marshal(TaskMessage)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
	}
	// publish to channel
	err = client.Publish(channelName, jsonMessage).Err()
	if nil != err {
		fmt.Printf("publish Error: %s", err.Error())
		return err
	}

	fmt.Println("published task to: ", channelName)
	return nil

}

//PublishArchiveTask marshals an archiveTask into json
// and publishes to the 'archive' redis channel
// where the archive request will be handled by a subscriber/worker
func PublishArchiveTask(client *redis.Client, archiveTask interface{}) error {

	//convert struct to json
	jsonMessage, err := json.Marshal(archiveTask)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
	}
	// publish to channel
	channelName := "archives"
	err = client.Publish(channelName, jsonMessage).Err()
	if nil != err {
		fmt.Printf("publish Error: %s", err.Error())
		return err
	}

	fmt.Println("published archive task")
	return nil

}

func PublishSourceBucketTask(client *redis.Client, bucket interface{}) error {
	jsonMessage, err := json.Marshal(bucket)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
	}
	// publish to channel
	channelName := "sourceBuckets"
	err = client.Publish(channelName, jsonMessage).Err()
	if err != nil {
		fmt.Printf("publish Error: %s", err.Error())
		return err
	}

	fmt.Println("published source bucket task")
	return nil
}

func PublishEmailTask(client *redis.Client, emailParts EmailParts) error {

	//convert struct to json
	jsonMessage, err := json.Marshal(emailParts)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
		return err
	}
	//publish to channel
	channelName := "emails"
	err = client.Publish(channelName, jsonMessage).Err()
	if nil != err {
		fmt.Printf("Publish Error: %s", err.Error())
		return err
	}

	fmt.Println("published email task")
	return nil

}
