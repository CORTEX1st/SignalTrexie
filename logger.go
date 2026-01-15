package main

import (
	"fmt"
	"os"
	"time"
)

func Info(msg string) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(os.Stderr, "%s [INFO] %s\n", timestamp, msg)
}

func Error(msg string) {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(os.Stderr, "%s [ERROR] %s\n", timestamp, msg)
}
