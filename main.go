package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func GetFeedEntries(url string) ([]Entry, error) {
	// client := &http.Client{}
	// req, err := http.NewRequest("GET", url, nil)
	// if err != nil {
	// 	return nil, err
	// }

	// req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36(KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")

	// resp, err := client.Do(req)
	// if err != nil {
	// 	return nil, err
	// }

	// defer resp.Body.Close()

	// byteValue, _ := ioutil.ReadAll(resp.Body)

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

type Request struct {
	URL string `json:"url"`
}

var client *mongo.Client
var ctx context.Context
var err error

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx = context.Background()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI((os.Getenv("MONGO_URI"))))
	if err != nil {
		log.Fatal("Cannot connect to mongo database ", err.Error())
	}

}
func main() {
	router := gin.Default()
	router.POST("/parse", Parsehandler)
	router.Run(":5000")

}

func Parsehandler(c *gin.Context) {
	var request Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	entries, err := GetFeedEntries(request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error while parsing the rss feed",
			"error":   err.Error(),
		})
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	for _, entry := range entries[2:] {
		collection.InsertOne(ctx, bson.M{
			"title":     entry.Title,
			"thumbnail": entry.Thumbnail.URL,
			"url":       entry.Link.Href,
		})

	}

	c.JSON(http.StatusOK, entries)
}
