package redisutil

import (
	"fmt"

	"github.com/go-redis/redis"
)

func InitRedisConnection(addr string) *redis.Client {

	rdbClient := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0, // redis' default database name
	})

	// check if redis db is reachable
	pong, err := rdbClient.Ping().Result()
	if (err != nil) || (pong != "PONG") {
		fmt.Println(err, " got: ", pong)
	}
	return rdbClient

}
