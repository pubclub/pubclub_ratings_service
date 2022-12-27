package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

type DynamoAPI struct {
	Db dynamodbiface.DynamoDBAPI
}

type Rating struct {
	RatingId     string `json:"RatingId"`
	CreationDate string `json:"CreationDate"`
	UserId       string `json:"UserId"`
	PlaceId      string `json:"PlaceId"`
}

var TableName string = "ratings-table"

func addRatingToDB(dyna DynamoAPI, rating Rating) (*dynamodb.PutItemOutput, error) {

	putItem := map[string]*dynamodb.AttributeValue{
		"RatingId":     {S: aws.String(rating.RatingId)},
		"CreationDate": {S: aws.String(rating.CreationDate)},
		"UserId":       {S: aws.String(rating.UserId)},
		"PlaceId":      {S: aws.String(rating.PlaceId)},
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

func healthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{"Status": "Available"})
}

func createRating(c *gin.Context) {
	var newRating Rating

	if err := c.BindJSON(&newRating); err != nil {
		return
	}

	fmt.Println(newRating)

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

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	router := gin.Default()
	router.GET("/healthcheck", healthCheck)
	router.POST("/rating", createRating)
	ginLambda = ginadapter.New(router)
	lambda.Start(Handler)
	// TODO: Create a removeRating function
	// TODO: Create a getRatings function
	//     - Maybe also getRatingsById function
}
