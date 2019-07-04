package dynamodb

import (
	"context"
	"errors"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	d "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/vectorhacker/eventing"
)

const (
	keyPrefix                  = "_"
	defaultRecordsPerPartition = 100
	defaultRangeKey            = "partition"
	defaultHashKey             = "key"
)

type store struct {
	dynamodb          *d.DynamoDB
	hashKey           string
	rangeKey          string
	tableName         string
	itemsPerPartition int
}

// New returns a new event store
func New(dynamodb *d.DynamoDB, tableName string, options ...StoreOption) eventing.Store {
	s := &store{
		dynamodb:          dynamodb,
		tableName:         tableName,
		itemsPerPartition: defaultRecordsPerPartition,
		hashKey:           defaultHashKey,
		rangeKey:          defaultHashKey,
	}

	// apply options
	for _, opt := range options {
		opt(s)
	}

	return s
}

func (s *store) checkIdempotency(ctx context.Context, id string, records []eventing.Record) error {

	if len(records) == 0 {
		return nil
	}

	version := records[len(records)-1].Version

	history, err := s.Load(ctx, id, 0, version)
	if err != nil {
		return err
	}

	if len(history) < len(records) {
		return errors.New(d.ErrCodeConditionalCheckFailedException)
	}

	recent := history[len(history)-len(records):]
	if !reflect.DeepEqual(recent, records) {
		return errors.New(d.ErrCodeConditionalCheckFailedException)
	}

	return nil
}

// Save implements the Store interface
func (s *store) Save(ctx context.Context, id string, records []eventing.Record) error {
	if len(records) == 0 {
		return nil
	}

	input, err := s.makeUpdateInput(id, records)
	if err != nil {
		return err
	}

	_, err = s.dynamodb.UpdateItemWithContext(ctx, input)
	if err != nil {
		if err, ok := err.(awserr.Error); ok {
			switch err.Code() {
			case d.ErrCodeConditionalCheckFailedException:
				return s.checkIdempotency(ctx, id, records)
			}
		}

		return err
	}

	return nil
}

// Load implements the Store interface
func (s *store) Load(ctx context.Context, id string, from, to int) ([]eventing.Record, error) {

	input, err := s.makeQueryInput(id, from, to)
	if err != nil {
		return nil, err
	}

	records := []eventing.Record{}
	err = s.dynamodb.QueryPagesWithContext(
		ctx,
		input,
		func(out *d.QueryOutput, done bool) bool {

			for _, item := range out.Items {
				records = append(
					records,
					filterRecords(from, to, extractRecords(item))...,
				)
			}

			return true
		},
	)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (s *store) makeQueryInput(id string, from, to int) (*d.QueryInput, error) {
	startPartition := selectPartition(from, s.itemsPerPartition)
	endPartition := selectPartition(to, s.itemsPerPartition)

	keys := expression.KeyEqual(expression.Key(s.hashKey), expression.Value(id)).
		And(expression.KeyGreaterThanEqual(expression.Key(s.rangeKey), expression.Value(startPartition)))

	if endPartition != eventing.EndOfStream {
		keys = keys.And(expression.KeyLessThanEqual(
			expression.Key(s.rangeKey),
			expression.Value(endPartition),
		))
	}

	expr, err := expression.NewBuilder().
		WithKeyCondition(keys).
		Build()

	if err != nil {
		return nil, err
	}

	input := &d.QueryInput{
		ConsistentRead:            aws.Bool(true),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Select:                    aws.String(d.SelectAllAttributes),
		TableName:                 aws.String(s.tableName),
	}

	return input, nil
}

func (s *store) makeUpdateInput(id string, records []eventing.Record) (*d.UpdateItemInput, error) {
	partition := selectPartition(records[0].Version, s.itemsPerPartition)

	conditions := []expression.ConditionBuilder{}
	update := expression.Add(expression.Name("revision"), expression.Value(1))
	for _, record := range records {
		key := keyFromVersion(record.Version)
		condition := expression.AttributeNotExists(expression.Name(key))
		conditions = append(conditions, condition)
		update = update.Set(expression.Name(key), expression.Value(record.Data))
	}

	condition := conditions[0]
	for _, c := range conditions {
		condition = condition.And(c)
	}

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		WithCondition(condition).
		Build()
	if err != nil {
		return nil, err
	}

	input := &d.UpdateItemInput{
		Key: map[string]*d.AttributeValue{
			s.hashKey:  {S: aws.String(id)},
			s.rangeKey: {N: aws.String(strconv.Itoa(partition))},
		},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       expr.Condition(),
		UpdateExpression:          expr.Update(),
		TableName:                 aws.String(s.tableName),
	}

	return input, nil
}

func selectPartition(version, items int) int {
	return version / items
}
