package main

import (
	"log"
	"time"
)

const (
	FocusDuration = 25 * time.Minute
	BreakDuration = 5 * time.Minute
)

func main() {
	log.Println("Hourglass service starting...")
	log.Println("Break Duration: ", BreakDuration)
}
