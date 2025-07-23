package common

import (
	"log"
	"os"
)

func AppLogger() *log.Logger {
	return log.New(os.Stdout, "app:", log.LstdFlags|log.Lshortfile)
}
