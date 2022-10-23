package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var channelAmqp *amqp.Channel

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("URI : ", os.Getenv("RABBITMQ_URI"))

	amqpConnection, err := amqp.Dial(os.Getenv("RABBITMQ_URI"))
	if err != nil {
		log.Fatal(err)
	}
	// Most operations happen on a channel.  If any error is returned on a
	// channel, the channel will no longer be valid, throw it away and try with
	// a different channel.
	channelAmqp, err = amqpConnection.Channel()
	if err != nil {
		log.Fatal(err)
	}
}

type Request struct {
	URL string `json:"url"`
}

func ParserHandler(c *gin.Context) {
	var request Request

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	data, _ := json.Marshal(request)
	err := channelAmqp.Publish(
		"",
		os.Getenv("RABBITMQ_QUEUE"),
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error while publishing to RabbitMQ",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "success",
	})

}

func main() {
	router := gin.Default()
	router.POST("/parse", ParserHandler)
	router.Run(":5000")

}
