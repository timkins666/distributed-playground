package main

import (
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestInitDB(t *testing.T) {
	t.Setenv("DB_TYPE", "POSTGRES")
	t.Setenv("POSTGRES_HOST", "localhost")

	redisClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	_, err := initDB(redisClient)
	if err == nil {
		t.Error("expected connection error in test environment")
	}
}