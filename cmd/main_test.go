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
		ratingId:     "1",
		creationDate: "2022-12-23",
		userId:       "1",
		placeId:      "1",
	}

	putItem := map[string]*dynamodb.AttributeValue{
		"RatingId":     {S: aws.String(rating.ratingId)},
		"CreationDate": {S: aws.String(rating.creationDate)},
		"UserId":       {S: aws.String(rating.userId)},
		"PlaceId":      {S: aws.String(rating.placeId)},
	}

	mock.ExpectPutItem().ToTable(TableName).WithItems(putItem)

	_, _ = addRatingToDB(dyna, rating)
}
