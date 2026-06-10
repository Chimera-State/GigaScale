package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{"localhost:6379"},
	})
	
	ctx := context.Background()
	err1 := rdb.SetArgs(ctx, "testlock", "user1", redis.SetArgs{
		Mode: "NX",
		TTL: 5 * time.Second,
	}).Err()
	fmt.Printf("First set err: %v\n", err1)

	err2 := rdb.SetArgs(ctx, "testlock", "user2", redis.SetArgs{
		Mode: "NX",
		TTL: 5 * time.Second,
	}).Err()
	fmt.Printf("Second set err: %v\n", err2)
	fmt.Printf("Is redis.Nil? %v\n", err2 == redis.Nil)
}
