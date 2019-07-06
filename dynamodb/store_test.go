package dynamodb

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/vectorhacker/eventing"
)

func TestDynamoDBStore(t *testing.T) {
	t.Parallel()

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		t.SkipNow()
		return
	}

	api, err := DynamoDB("us-east-1", endpoint)
	assert.Nil(t, err)

	TempTable(t, api, func(tableName string) {
		ctx := context.Background()
		store := New(api, tableName)
		assert.Nil(t, err)

		aggregateID := "abc"
		history := []eventing.Record{}

		for i := 0; i < 100; i++ {
			history = append(history, eventing.Record{
				Data:    []byte("a"),
				Version: i + 1,
			})
		}

		err = store.Save(ctx, aggregateID, history)
		assert.Nil(t, err)

		found, err := store.Load(ctx, aggregateID, 0, 0)
		assert.Nil(t, err)
		assert.Equal(t, history, found)
		assert.Len(t, found, len(history))
	})
}

var (
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func TempTable(t *testing.T, api *dynamodb.DynamoDB, fn func(tableName string)) {
	// Create a temporary table for use during this test
	//
	now := strconv.FormatInt(time.Now().UnixNano(), 36)
	random := strconv.FormatInt(int64(r.Int31()), 36)
	tableName := "tmp-" + now + "-" + random

	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String(defaultHashKey),
				AttributeType: aws.String("S"),
			},
			{
				AttributeName: aws.String(defaultRangeKey),
				AttributeType: aws.String("N"),
			},
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String(defaultHashKey),
				KeyType:       aws.String("HASH"),
			},
			{
				AttributeName: aws.String(defaultRangeKey),
				KeyType:       aws.String("RANGE"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(1),
			WriteCapacityUnits: aws.Int64(1),
		},
		StreamSpecification: &dynamodb.StreamSpecification{
			StreamEnabled:  aws.Bool(true),
			StreamViewType: aws.String("NEW_AND_OLD_IMAGES"),
		},
	}

	_, err := api.CreateTable(input)
	assert.Nil(t, err)
	defer func() {
		_, err := api.DeleteTable(&dynamodb.DeleteTableInput{TableName: aws.String(tableName)})
		assert.Nil(t, err)
	}()

	fn(tableName)
}
