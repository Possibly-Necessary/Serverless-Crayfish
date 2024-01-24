// Go script that publishes sub-populations of crayfish to RabbitMQ's message broker
// using the default exchange

package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"

	"github.com/streadway/amqp"
)

// For the RabbitMQ part
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

func initializePopulation(N, k int, ch *amqp.Channel, queueName string) error { // Instead of returning ([]byte, error)

	// This part is going to be provided by the benchmarks library
	lb := []float64{-100.0}
	ub := []float64{100.0}

	dim := 3

	// Initialize the population N x Dim matrix, X
	X := make([][]float64, N)
	for i := 0; i < N; i++ {
		X[i] = make([]float64, dim)
	}

	for i := range X {
		for j := range X[i] {
			X[i][j] = rand.Float64()*(ub[0]-lb[0]) + lb[0]
		}
	}

	// Split the population based on k
	totalSize := len(X)
	baseSubPopSize := totalSize / k
	remainder := totalSize % k

	Xsub := make([][][]float64, k)

	startIndex := 0
	//subPopCount := 0

	for i := 0; i < k; i++ {
		subPopSize := baseSubPopSize
		if remainder > 0 { // In case the division is not even
			subPopSize++ // Add one of the remaining individuals to this sub-population
			remainder--
		}
		Xsub[i] = X[startIndex : startIndex+subPopSize]
		startIndex += subPopSize

		msg := Message{
			SubPopulation: Xsub[i],
			Workers:       k,
			F:             "F6",
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(msg)
		if err != nil {
			return fmt.Errorf("Failed to encode Crayfish data.")
		}

		err = ch.Publish(
			"",        // exchange
			queueName, // routing key (was q.Name)
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/octet-stream",
				Body:        buf.Bytes(),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent message %d", i)

		//subPopCount++
	}

	return nil
}

func main() {
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

	err = initializePopulation(20, 4, ch, q.Name)
	failOnError(err, "Failed to initialize population. ")

}
