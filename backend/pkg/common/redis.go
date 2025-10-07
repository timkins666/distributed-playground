package common

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisEntityKey string

var (
	RedisKeyUser         RedisEntityKey = "user"
	RedisKeyUserAccounts RedisEntityKey = "userAccounts"
)

func RedisKey(entityKey RedisEntityKey, id string) string {
	return fmt.Sprintf("%s:%s", entityKey, id)
}

func NewRedisClient() (*redis.Client, error) {
	addr := "redis:6379" // TODO: env/pass in
	client := redis.NewClient(&redis.Options{
		Addr: addr,
		// Password: "",
		// DB:       0, // TODO: ??
	})

	var err error
	for i := 0; i < 10; i++ {
		if err = client.Ping(context.Background()).Err(); err == nil {
			return client, nil
		}
	}
	return nil, fmt.Errorf("failed to ping redis: %s", err.Error())
}
