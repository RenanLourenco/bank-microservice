package event

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func declareExchange(ch *amqp.Channel) error {
	//declaring exchanges

	errExchangeAuth := ch.ExchangeDeclare(
		"auth_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if errExchangeAuth != nil {
		log.Println("Error exchanging auth")
		return errExchangeAuth
	}

	errExchangeTransaction := ch.ExchangeDeclare(
		"transaction_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if errExchangeTransaction != nil {
		log.Println("Error exchanging transaction")
		return errExchangeTransaction
	}

	errExchangeDeposit := ch.ExchangeDeclare(
		"deposit_topic",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if errExchangeDeposit != nil {
		log.Println("Error exchanging deposit")
		return errExchangeDeposit
	}

	return nil
}

func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"queue", // name?
		false, // durable?
		false, // delete when unused?
		true, // exclusive?
		false, // no-wait
		nil, // arguments
	)
}

func declareAuthQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"auth_queue",
		false,
		false,
		true,
		false,
		nil,
	)
}

func declareTransactionQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"transaction_queue",
		false,
		false,
		true,
		false,
		nil,
	)
}

func declareDepositQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"deposit_queue",
		false,
		false,
		true,
		false,
		nil,
	)
}
