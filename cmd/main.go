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
	PlaceRating  string `json:"PlaceRating"`
}

var TableName string = "ratings-table"

func addRatingToDB(dyna DynamoAPI, rating Rating) (*dynamodb.PutItemOutput, error) {

	putItem := map[string]*dynamodb.AttributeValue{
		"RatingId":     {S: aws.String(rating.RatingId)},
		"CreationDate": {S: aws.String(rating.CreationDate)},
		"UserId":       {S: aws.String(rating.UserId)},
		"PlaceId":      {S: aws.String(rating.PlaceId)},
		"PlaceRating":  {N: aws.String(rating.PlaceRating)},
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

func removeRating(dyna DynamoAPI, RatingId string) error {

	deleteItem := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"RatingId": {S: aws.String(RatingId)},
		},
		TableName: aws.String(TableName),
	}
	_, err := dyna.Db.DeleteItem(deleteItem)
	if err != nil {
		log.Fatalf("Unable to delete item with rating ID: %s", RatingId)
		return err
	}

	return nil
}

func getRatingById(dyna DynamoAPI, placeId string) ([]*string, error) {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"PlaceId": {
				N: aws.String(placeId),
			},
		},
		TableName: aws.String(TableName),
	}

	response, err := dyna.Db.GetItem(input)
	if err != nil {
		return nil, err
	}

	rating := response.Item["Ratings"].NS

	return rating, nil
}

func healthCheck(c *gin.Context) {
	c.String(http.StatusOK, "alive")
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

func setUpRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/healthcheck", healthCheck)
	router.POST("/rating", createRating)

	return router
}

func main() {
	router := setUpRouter()
	ginLambda = ginadapter.New(router)
	lambda.Start(Handler)
	// TODO: Create a removeRating function
	// TODO: Create a getRatings function
	//     - Maybe also getRatingsById function
}
