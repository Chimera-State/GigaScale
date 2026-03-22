package redisclient

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient redis.UniversalClient // Redis bağlantısı

func InitRedisCluster(){
	clusterAddrs := []string{
		"redis-node-1:6379", "redis-node-2:6379", "redis-node-3:6379",
		"redis-node-4:6379", "redis-node-5:6379", "redis-node-6:6379",
	}

	RedisClient = redis.NewClusterClient(&redis.ClusterOptions{ 
		Addrs:        clusterAddrs,
		MaxRedirects: 8,
		ReadOnly:     false,
	})

	log.Println("Redis Cluster başarıyla oluşturuldu")

}

func HealthCheck(ctx context.Context) {
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis bağlantı hatası:", err)
	} 
	log.Println("Redis bağlantısı başarılı")
}