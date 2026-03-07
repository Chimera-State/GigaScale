package redisclient

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client // Redis bağlantısı

func NewRedisClient(){
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" { 
		addr = "localhost:6379" // Eğer REDIS_ADDR ortam değişkeni yoksa default olarak localhost:6379 kullanılır.
	}

	RedisClient = redis.NewClient(&redis.Options{ 
		Addr: addr, 
		PoolSize: 10,
		MinIdleConns: 5,
		MaxRetries: 3,
	})

	log.Println("Redis'e başarıyla oluşturuldu:",addr)


}

func HealthCheck(ctx context.Context) {
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatal("Redis bağlantı hatası:", err)
	} 
	log.Println("Redis bağlantısı başarılı")
}