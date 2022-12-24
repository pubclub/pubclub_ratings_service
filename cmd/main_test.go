package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	dynamock "github.com/gusaul/go-dynamock"
)

var mock *dynamock.DynaMock

func TestAddRatingToDB(t *testing.T) {

	var dyna DynamoAPI
	dyna.Db, mock = dynamock.New()

	rating := Rating{
		RatingId:     "1",
		CreationDate: "2022-12-23",
		UserId:       "1",
		PlaceId:      "1",
	}

	putItem := map[string]*dynamodb.AttributeValue{
		"RatingId":     {S: aws.String(rating.RatingId)},
		"CreationDate": {S: aws.String(rating.CreationDate)},
		"UserId":       {S: aws.String(rating.UserId)},
		"PlaceId":      {S: aws.String(rating.PlaceId)},
	}

	mock.ExpectPutItem().ToTable(TableName).WithItems(putItem)

	_, _ = addRatingToDB(dyna, rating)
}
