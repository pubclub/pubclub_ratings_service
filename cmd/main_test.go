package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-playground/assert/v2"
	dynamock "github.com/gusaul/go-dynamock"
)

var mock *dynamock.DynaMock

func createPutItem(rating Rating) map[string]*dynamodb.AttributeValue {

	putItem := map[string]*dynamodb.AttributeValue{
		"RatingId":     {S: aws.String(rating.RatingId)},
		"CreationDate": {S: aws.String(rating.CreationDate)},
		"UserId":       {S: aws.String(rating.UserId)},
		"PlaceId":      {S: aws.String(rating.PlaceId)},
		"PlaceRating":  {N: aws.String(rating.PlaceRating)},
	}

	return putItem
}

func TestAddRatingToDB(t *testing.T) {

	var dyna DynamoAPI
	dyna.Db, mock = dynamock.New()

	rating := Rating{
		RatingId:     "1",
		CreationDate: "2022-12-23",
		UserId:       "1",
		PlaceId:      "1",
		PlaceRating:  "4",
	}

	putItem := createPutItem(rating)

	mock.ExpectPutItem().ToTable(TableName).WithItems(putItem)

	_, _ = addRatingToDB(dyna, rating)
}

func TestRemoveRating(t *testing.T) {

	var dyna DynamoAPI
	dyna.Db, mock = dynamock.New()

	removeRatingId := "1"
	removeCreationDate := "2022-12-23"

	rating := Rating{
		RatingId:     "1",
		CreationDate: "2022-12-23",
		UserId:       "1",
		PlaceId:      "1",
		PlaceRating:  "4",
	}

	putItem := createPutItem(rating)

	// Need to ensure the mock database expects an item to be uploaded
	// so can later be removed
	mock.ExpectPutItem().ToTable(TableName).WithItems(putItem)
	_, _ = addRatingToDB(dyna, rating)

	expectKey := map[string]*dynamodb.AttributeValue{
		"RatingId": {
			S: aws.String(removeRatingId),
		},
		"CreationDate": {
			S: aws.String(removeCreationDate),
		},
	}

	mock.ExpectDeleteItem().ToTable(TableName).WithKeys(expectKey)

	error := removeRatingFromDB(dyna, removeRatingId, removeCreationDate)

	if error != nil {
		t.Errorf("Unable to delete item")
	}

}

func TestGetRatingById(t *testing.T) {

	var dyna DynamoAPI
	dyna.Db, mock = dynamock.New()

	getRatingPlaceId := "1"

	firstRating := Rating{
		RatingId:     "1",
		CreationDate: "2023-01-03",
		UserId:       "1",
		PlaceId:      getRatingPlaceId,
		PlaceRating:  "4",
	}

	secondRating := Rating{
		RatingId:     "2",
		CreationDate: "2023-01-03",
		UserId:       "2",
		PlaceId:      getRatingPlaceId,
		PlaceRating:  "3",
	}

	firstPutItem := createPutItem(firstRating)
	secondPutItem := createPutItem(secondRating)

	// Need to ensure the mock database expects items to be uploaded
	// so ratings can be grabbed
	mock.ExpectPutItem().ToTable(TableName).WithItems(firstPutItem)
	mock.ExpectPutItem().ToTable(TableName).WithItems(secondPutItem)

	_, _ = addRatingToDB(dyna, firstRating)
	_, _ = addRatingToDB(dyna, secondRating)

	expectKey := map[string]*dynamodb.AttributeValue{
		"id": {
			N: aws.String(getRatingPlaceId),
		},
	}

	expectedRating := []string{firstRating.PlaceRating, secondRating.PlaceRating}
	expectedRatingSet := aws.StringSlice(expectedRating)

	result := dynamodb.GetItemOutput{
		Item: map[string]*dynamodb.AttributeValue{
			"Ratings": {
				NS: expectedRatingSet,
			},
		},
	}

	mock.ExpectGetItem().ToTable(TableName).WithKeys(expectKey).WillReturns(result)
	actualRating, _ := getRatingById(dyna, getRatingPlaceId)
	for ratingIndex, rating := range actualRating {
		if *rating != expectedRating[ratingIndex] {
			t.Errorf("Test Fail")
		}
	}
}

func TestHealthCheck(t *testing.T) {
	router := setUpRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "alive", w.Body.String())
}
