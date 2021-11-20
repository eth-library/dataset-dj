package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
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

// for testing purposes create a dummy archive based on the variable i
func createMetaArchive(i int) metaArchive {

	newUid := uuid.New().String()[:8]

	file1 := fmt.Sprintf("image_%v.png", i)
	file2 := fmt.Sprintf("image_%v.png", i+1)
	newMetaArch := metaArchive{
		ID:          newUid,
		Files:       []string{file1, file2},
		TimeCreated: time.Now().String(),
		TimeUpdated: "",
		Status:      "opened",
	}

	return newMetaArch
}

func publishToChannel(i int) {

	channelName := "default"
	var newMetaArch metaArchive

	newMetaArch = createMetaArchive(i)
	jsonMessage, err := json.Marshal(newMetaArch)
	if err != nil {
		fmt.Println("error marshalling json: ", err)
	}

	err = rdbClient.Publish(channelName, jsonMessage).Err()
	if nil != err {
		fmt.Printf("Publish Error: %s", err.Error())
	}
	fmt.Println(i, "    published task: ", newMetaArch.ID, "  ", newMetaArch.Files)

}

func main() {

	fmt.Println("started task publisher")

	rdbClient = initRedisConnection()

	// add a record as a `Key`  `Value` pair
	// key has to be a unique string
	// value can be single value for type
	key := "1234"
	value := "a super task"
	err := rdbClient.Set("1234", value, 0).Err()
	// handle the error
	if err != nil {
		fmt.Println(err)
	}

	val, err := rdbClient.Get(key).Result()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("response for %v:, %v \n", key, val)

	key2 := "import_value_1"
	val2 := 42
	err = rdbClient.Set(key2, val2, 0).Err()
	// handle the error
	if err != nil {
		fmt.Println(err)
	}

	val2_returned, err := rdbClient.Get(key2).Result()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("response for %v: %v \n", key2, val2_returned)

	//create a new metaArchive
	newMetaArch := metaArchive{
		ID:          "asd123",
		Files:       []string{"file1.png", "file2.png"},
		TimeCreated: time.Now().String(),
		TimeUpdated: "",
		Status:      "opened",
	}
	fmt.Println("created task: ", newMetaArch)

	// need to convert struct to json using json.Marshal before sending to Redis
	json, err := json.Marshal(newMetaArch)

	err = rdbClient.Set(newMetaArch.ID, json, 0).Err()
	if err != nil {
		fmt.Println("error setting task: ", err)
	}
	fmt.Println("set task: ", newMetaArch.ID)

	// get the task back from Redis
	val, err = rdbClient.Get(newMetaArch.ID).Result()
	if err != nil {
		fmt.Println("error getting task: ", err)
	}
	fmt.Println("got returned task: ", val)

	var i int
	i, err = strconv.Atoi(os.Args[1])
	publishToChannel(i)

}
