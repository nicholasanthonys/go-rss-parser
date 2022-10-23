package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err.Error())
	}
	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}
	defer amqpConnection.Close()

	channelAmqp, _ := amqpConnection.Channel()
	defer channelAmqp.Close()

	forever := make(chan bool)

	msgs, err := channelAmqp.Consume(
		os.Getenv("RABBITMQ_QUEUE"),
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
