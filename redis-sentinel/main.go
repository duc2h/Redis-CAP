package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {

	// Redis will auto failover to the new master if the current master is down.
	rdb := redis.NewFailoverClusterClient(&redis.FailoverOptions{
		MasterName:    "mymaster",
		SentinelAddrs: []string{"redis-sentinel1:26379", "redis-sentinel2:26380", "redis-sentinel3:26381"},
	})

	// set a key
	err := rdb.Set(context.Background(), "key", "value111", 500*time.Second).Err()
	if err != nil {
		fmt.Println("Error setting key: ", err)
	}

	// get a key
	val, err := rdb.Get(context.Background(), "key").Result()
	if err != nil {
		fmt.Println("Error getting key: ", err)
	} else {
		fmt.Println("key: ", val)
	}
}
