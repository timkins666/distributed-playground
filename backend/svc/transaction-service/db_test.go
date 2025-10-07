package main

import (
	"testing"
)

func TestInitDB(t *testing.T) {
	t.Setenv("DB_TYPE", "POSTGRES")
	t.Setenv("POSTGRES_HOST", "localhost")

	_, err := initDB()
	if err == nil {
		t.Error("expected connection error in test environment")
	}
}