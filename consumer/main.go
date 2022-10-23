package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Request struct {
	URL string `json:"url"`
}

type Feed struct {
	Entries []Entry `xml:"entry"`
}
type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`

	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`

	Title string `xml:"title"`
}

var mongoClient *mongo.Client
var ctx context.Context
var err error

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx = context.Background()

	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI((os.Getenv("MONGO_URI"))))
	if err != nil {
		log.Fatal("Cannot connect to mongo database ", err.Error())
	}

}

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

			var request Request

			json.Unmarshal(d.Body, &request)

			log.Println("RSS URL : ", request.URL)

			entries, _ := GetFeedEntries(request.URL)

			collection := mongoClient.Database(os.Getenv("MONGO_DATABASE")).
				Collection("recipes")
			for _, entry := range entries[2:] {
				collection.InsertOne(ctx, bson.M{
					"title" : entry.Title,
					"thumbnail" : entry.Thumbnail.URL,
					"url" : entry.Link.Href,
				})

			}

		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func GetFeedEntries(url string) ([]Entry, error) {
	xmlFile, err := os.Open("rss.xml")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened rss.xml")
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)

	var feed Feed
	xml.Unmarshal(byteValue, &feed)

	fmt.Print("feed entries : ", feed.Entries)

	return feed.Entries, nil
}
