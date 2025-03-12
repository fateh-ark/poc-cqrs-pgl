package main

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	var err error

	// Rabbit MQ Setup
	rabbitConn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rabbitConn.Close()

	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer rabbitChan.Close()

	err = rabbitChan.ExchangeDeclare(
		"book_events", // Exchange name
		"topic",       // Exchange type (topic)
		true,          // Durable
		false,         // Auto-deleted
		false,         // Internal
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	queue, err := rabbitChan.QueueDeclare(
		"",    // Queue name (auto-generated)
		false, // Durable
		true,  // Delete when unused
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	err = rabbitChan.QueueBind(
		queue.Name,    // Queue name
		"book.*",      // Routing key pattern (e.g., "book.created", "book.updated")
		"book_events", // Exchange
		false,         // No-wait
		nil,           // Arguments
	)
	if err != nil {
		log.Fatal("Failed to bind a queue:", err)
	}

	msgs, err := rabbitChan.Consume(
		queue.Name, // Queue
		"",         // Consumer
		true,       // Auto-ack
		false,      // Exclusive
		false,      // No-local
		false,      // No-wait
		nil,        // Args
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var event map[string]interface{}
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Println("Error unmarshalling message:", err)
				continue
			}
			log.Printf("Received event (%s): %v", d.RoutingKey, event)
		}
	}()

	log.Printf(" [*] Waiting for messages")
	<-forever // run the script forever
}
