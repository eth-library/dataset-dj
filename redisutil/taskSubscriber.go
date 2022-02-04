package redisutil

import "github.com/go-redis/redis"

func SubscribeToRedisChannel(rdbClient *redis.Client, handlers map[string]interface{}) {

	//can subscribe to messages from multiple channels
	subscriber := rdbClient.Subscribe("archives", "emails", "sourceBuckets")
	defer subscriber.Close()
	channel := subscriber.Channel()

	for msg := range channel {
		handlers[msg.Channel].(func(string))(msg.Payload)
	}
}
