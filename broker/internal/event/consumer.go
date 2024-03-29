package event

import (
	transactions_proto "broker/internal/transactions-proto"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()

	if err != nil {
		return Consumer{}, nil
	}

	return consumer, nil

}

func (c *Consumer) setup() error {
	channel, err := c.conn.Channel()

	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type PayloadTransaction struct {
	Name string                          `json:"name"`
	Data *transactions_proto.Transaction `json:"data"`
}
type PayloadBalance struct {
	Name string                      `json:"name"`
	Data *transactions_proto.Balance `json:"data"`
}

func (c *Consumer) ListenTransaction(topics []string) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := declareTransactionQueue(ch)
	if err != nil {
		return err
	}

	for _, s := range topics {
		err = ch.QueueBind(
			q.Name,
			s,
			"transaction_topic",
			false,
			nil,
		)
		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	f := make(chan bool)

	go func() {
		for d := range messages {
			var payload PayloadTransaction
			_ = json.Unmarshal(d.Body, &payload)
			go handleTransaction(payload)
		}
	}()

	fmt.Printf("Waiting for message [Exchange, Queue] [transaction_topic, %s]", q.Name)
	<-f

	return nil

}

func handleTransaction(payload PayloadTransaction) {
	log.Println("Handling transaction message")
	input := payload.Data
	err := transactionEvent(input)
	if err != nil {
		log.Println(err)
	}
}

func handleBalance(payloadBalance PayloadBalance) {
	log.Println("Handling transaction message")
	input := payloadBalance.Data
	err := balanceEvent(input)
	if err != nil {
		log.Println(err)
	}
}

func transactionEvent(payload *transactions_proto.Transaction) error {

	conn, err := grpc.Dial("transaction:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	c := transactions_proto.NewTransactionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.CreateTransaction(ctx, &transactions_proto.TransactionRequest{
		TransactionEntry: &transactions_proto.Transaction{
			Value:      payload.Value,
			FromUserId: payload.FromUserId,
			ToUserId:   payload.FromUserId,
		},
	})
	if err != nil {
		return nil
	}
	return nil
}

func balanceEvent(payload *transactions_proto.Balance) error {
	conn, err := grpc.Dial("transaction:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return err
	}

	c := transactions_proto.NewTransactionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = c.CreateBalance(ctx, &transactions_proto.BalanceRequest{
		BalanceEntry: &transactions_proto.Balance{
			UserId:  payload.UserId,
			Balance: payload.Balance,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

// func (c *Consumer) ListenAuth(topics []string) error {
// 	ch, err := c.conn.Channel()
// 	if err != nil {
// 		return err
// 	}

// 	defer ch.Close()

// 	q, err := declareAuthQueue(ch)
// 	if err != nil {
// 		return err
// 	}

// 	for _, s := range topics {
// 		err = ch.QueueBind(
// 			q.Name,
// 			s,
// 			"auth_topic",
// 			false,
// 			nil,
// 		)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	messages, err := ch.Consume(
// 		q.Name,
// 		"",
// 		true,
// 		false,
// 		false,
// 		false,
// 		nil,
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	f := make(chan bool)

// 	go func ()  {
// 		for d := range messages {
// 			var payload Payload
// 			_ = json.Unmarshal(d.Body, &payload)
// 			if payload.Name == "balance"{
// 				var balancePayload PayloadBalance
// 				_ = json.Unmarshal(d.Body, &balancePayload)
// 				go handleBalance(balancePayload)
// 			}else{
// 				go handleTransaction(payload)
// 			}
// 		}
// 	}()

// 	fmt.Printf("Waiting for message [Exchange, Queue] [auth_topic, %s]", q.Name)
// 	<-f

// 	return nil

// }
