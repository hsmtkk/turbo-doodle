package main

// Import Go and NATS packages
import (
    "log"
    "runtime"

    "github.com/nats-io/nats.go"
)

func main() {

    // Create server connection
    natsConnection, _ := nats.Connect("nats://nats:4222")
    log.Println("Connected")

    // Subscribe to subject
    log.Printf("Subscribing to subject 'test'\n")
    natsConnection.Subscribe("test", func(msg *nats.Msg) {

        // Handle the message
        log.Printf("Received message '%s\n", string(msg.Data)+"'")
    })

    // Keep the connection alive
    runtime.Goexit()
}
