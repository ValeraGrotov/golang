package main
import (
	"bufio"
	"strings"
	"github.com/go-redis/redis"
)
func main() {	
	redisdb := redis.NewClient(&redis.Options{ Addr: "addr:6379" })
	pubsub := redisdb.Subscribe("chat")
	_, err := pubsub.Receive()
	if err != nil { println(err); return }
	ch := pubsub.Channel()
	go func() {
		for msg := range ch {
			if !strings.Contains(msg.Payload, "2:BIG") {
				array := strings.Split(msg.Payload, ":BIG")
				println(array[1])
			}
		}
	}()
	for {
		var message string
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() { message = scanner.Text() }
		err = redisdb.Publish("chat", "2:BIG"+message).Err()
		if err != nil { println(err); return }
	}
}
