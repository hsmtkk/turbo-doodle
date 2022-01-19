package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	conn, err := nats.Connect("nats://192.168.11.13:4222")
	if err != nil {
		log.Fatalf("failed to connect NATS; %s", err)
	}
	defer conn.Close()

	ch := make(chan *nats.Msg)
	sub, err := conn.ChanSubscribe("test", ch)
	if err != nil {
		log.Fatalf("failed to subscribe channel; %s", err)
	}
	defer sub.Unsubscribe()
	for msg := range ch {
		fmt.Println(string(msg.Data))
	}
}
