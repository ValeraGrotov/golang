package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-redis/redis"
)

var redisdb *redis.Client

func init() {
	redisdb = redis.NewClient(&redis.Options{
		Addr:     "addr:6379",
	})
}

func main() {
	pubsub := redisdb.Subscribe("chat")
	_, err := pubsub.Receive()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	ch := pubsub.Channel()
  
	go func() {
		for msg := range ch {
			if !strings.Contains(msg.Payload, "2:BIG") {
				array := strings.Split(msg.Payload, ":BIG")
				message := array[1]
				fmt.Println(message)
			}
		}
	}()

	for {
		var message string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			message = scanner.Text()
		}

		err = redisdb.Publish("chat", "2:BIG"+message).Err()
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}
