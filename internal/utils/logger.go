package utils

import (
	"log"
	"os"
)

type Logger interface {
	Printf(s string, v ...any)
	Println(v ...any)
}

func InitLogger(prefix string) Logger {
	return log.New(os.Stdout, prefix, log.LstdFlags)
}
