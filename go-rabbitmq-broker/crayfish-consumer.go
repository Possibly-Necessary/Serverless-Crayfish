package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"sync"

	"github.com/streadway/amqp"
)

type Message struct {
	SubPopulation [][]float64
	Workers       int
	F             string // Function name
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Try this: amqp.Dial("amqp://guest:guest@rabbitmq-dev-hostname:5672/")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"testQueue", // queue name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	// Assume k is known (or somehow we were able to magically get it from some function)
	k := 4

	var wg sync.WaitGroup
	for i := 0; i < k; i++ {
		wg.Add(1)
		go func(id int) { // Using Goroutine
			defer wg.Done()
			for msg := range msgs {
				var message Message
				buf := bytes.NewBuffer(msg.Body)
				dec := gob.NewDecoder(buf) // Decode data
				if err := dec.Decode(&message); err != nil {
					log.Printf("Goroutine %d failed  to decode message %s", id, err)
					continue
				}
				//log.Printf("Goroutine %d received a message: %+v", id, message)
				//log.Printf("SubPopulation: %+v", message.SubPopulation)
				//log.Printf("Workers: %d", message.Workers)
				//log.Printf("Function Name: %s", message.F)
				log.Printf("Goroutine %d received a message: Function: %s, SubPopulation: %+v\n",
					id, message.F, message.SubPopulation) //change message.SubPopulation[0]
				return // Exit goroutine after processing one message
			}
		}(i)
	}
	wg.Wait()
}
