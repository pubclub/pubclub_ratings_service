package main

import (
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/gin-gonic/gin"
)

type DynamoAPI struct {
	Db dynamodbiface.DynamoDBAPI
}

type Rating struct {
	ratingId     string `json:"RatingId"`
	creationDate string `json:"CreationDate"`
	userId       string `json:"UserId"`
	placeId      string `json:"PlaceId"`
}

var TableName string = "ratings-table"

func addRatingToDB(dyna DynamoAPI, rating Rating) (*dynamodb.PutItemOutput, error) {

	putItem := map[string]*dynamodb.AttributeValue{
		"RatingId":     {S: aws.String(rating.ratingId)},
		"CreationDate": {S: aws.String(rating.creationDate)},
		"UserId":       {S: aws.String(rating.userId)},
		"PlaceId":      {S: aws.String(rating.placeId)},
	}

	input := &dynamodb.PutItemInput{
		Item:      putItem,
		TableName: aws.String(TableName),
	}

	output, err := dyna.Db.PutItem(input)
	if err != nil {
		log.Fatalf("Got error calling PutItem: %s", err)
		return nil, err
	}
	return output, nil
}

func createRating(c *gin.Context) {
	var newRating Rating

	if err := c.BindJSON(&newRating); err != nil {
		return
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)
	var dyna DynamoAPI
	dyna.Db = svc

	_, err := addRatingToDB(dyna, newRating)
	if err != nil {
		panic(err)
	}

	c.IndentedJSON(http.StatusCreated, newRating)
}

func main() {
	router := gin.Default()
	router.POST("/rating", createRating)
	router.Run("localhost:8080")
}
