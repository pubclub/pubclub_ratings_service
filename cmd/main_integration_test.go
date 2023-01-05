//go:build integration

package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/go-playground/assert/v2"
)

var uploadString string = `{"RatingId": "9999", "CreationDate": "2022-01-01", "UserId": "9999", "PlaceId": "9999", "PlaceRating": "5"}`

func createServiceClient() *dynamodb.DynamoDB {

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return dynamodb.New(sess)
}

func createExpression(ratingId string) (expression.Expression, error) {
	filt := expression.Name("RatingId").Equal(expression.Value(ratingId))
	proj := expression.NamesList(
		expression.Name("RatingId"),
		expression.Name("CreationDate"),
		expression.Name("UserId"),
		expression.Name("PlaceId"),
	)
	expr, err := expression.NewBuilder().WithFilter(filt).WithProjection(proj).Build()
	if err != nil {
		log.Fatalf("Got error building expression: %s", err)
		return expr, err
	}
	return expr, nil
}

func scanTable(svc *dynamodb.DynamoDB, expr expression.Expression) (*dynamodb.ScanOutput, error) {

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(TableName),
	}

	// Make the DynamoDB Query API call
	result, err := svc.Scan(params)
	if err != nil {
		log.Fatalf("Query API call failed: %s", err)
		return nil, err
	}

	return result, err
}

// Deletes a previously created record using the global "9999" ID
func DeleteRecord(svc *dynamodb.DynamoDB, ratingId string, creationDate string) error {
	var dyna DynamoAPI
	dyna.Db = svc
	err := removeRating(dyna, ratingId, creationDate)
	if err != nil {
		panic("Unable to delete record")
	}
	return err
}

func TestCreateAndDeleteRating(t *testing.T) {

	router := setUpRouter()

	w := httptest.NewRecorder()
	reader := strings.NewReader(uploadString)
	req, _ := http.NewRequest("POST", "/rating", reader)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	router.ServeHTTP(w, req)

	assert.Equal(t, 201, w.Code)

	// Checks an item has actually been uploaded to DynamoDB
	svc := createServiceClient()

	ratingId := "9999" // Always use 9999 for testing purposes
	creationDate := "2022-01-01"
	expr, err := createExpression(ratingId)
	if err != nil {
		t.Error("Unable to build expression")
	}
	result, err := scanTable(svc, expr)
	if err != nil {
		t.Error("Scanning table failed")
	}

	items := result.Items
	assert.Equal(t, 1, len(items))

	rating := Rating{}
	err = dynamodbattribute.UnmarshalMap(items[0], &rating)

	assert.Equal(t, "9999", rating.RatingId)
	assert.Equal(t, "2022-01-01", rating.CreationDate)
	assert.Equal(t, "9999", rating.UserId)
	assert.Equal(t, "9999", rating.PlaceId)

	err = DeleteRecord(svc, ratingId, creationDate)

	deleteResult, err := scanTable(svc, expr)
	if err != nil {
		t.Error("Scanning table failed")
	}

	deleteItems := deleteResult.Items
	assert.Equal(t, 0, len(deleteItems))

}
