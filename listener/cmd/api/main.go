package main

import (
	"fmt"
	"listener/event"
	"log"
	"math"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const webPort = "80"

type Config struct {
	Rabbit *amqp.Connection
}

func main() {
	rabbitCon, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	defer rabbitCon.Close()

	//start listening for messages
	log.Println("Listening for and consuming RabbitMQ messages...")

	//create consumer
	consumer, err := event.NewConsumer(rabbitCon)
	if err != nil {
		panic(err)
	}

	//watch the queue and consume events

	//auth queue
	//err = consumer.ListenAuth([]string{"auth.INFO", "auth.WARNING", "auth.ERROR"})

	if err != nil {
		log.Println(err)
	}

	//deposit queue
	go consumer.ListenDeposit([]string{"deposit.INFO", "deposit.WARNING", "deposit.ERROR"})

	// transaction queue
	err = consumer.ListenTransaction([]string{"transaction.INFO", "transaction.WARNING", "transaction.ERROR"})

	if err != nil {
		log.Println(err)
	}
}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backoff = 1 * time.Second
	var connection *amqp.Connection

	// dont continue until rabbitmq is ready
	for {
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready")
			counts++
		} else {
			log.Println("Connected to RabbitMQ")
			connection = c
			break
		}

		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		backoff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off...")
		time.Sleep(backoff)
		continue
	}

	return connection, nil
}
