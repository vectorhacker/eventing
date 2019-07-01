package dynamodb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	d "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/vectorhacker/eventing"
)

const (
	keyPrefix = "_"
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
		dynamodb:  dynamodb,
		tableName: tableName,
	}

	// apply options
	for _, opt := range options {
		opt(s)
	}

	return s
}

func (s *store) Load(
	ctx context.Context,
	id string,
	from, to int,
) ([]eventing.Record, error) {

	input, err := s.makeQueryInput(id, from, to)
	if err != nil {
		return nil, err
	}

	records := []eventing.Record{}
	err = s.dynamodb.QueryPagesWithContext(ctx, input, func(out *d.QueryOutput, done bool) bool {
		for _, item := range out.Items {
			records = append(records, filterRecords(from, to, extractRecords(item))...)
		}

		return *out.Count > *out.ScannedCount
	})
	if err != nil {
		return nil, err
	}

	return records, nil
}

func extractRecords(item map[string]*d.AttributeValue) []eventing.Record {
	records := []eventing.Record{}
	for key, value := range item {
		if isKey(key) {
			v := versionFromKey(key)

			records = append(records, eventing.Record{
				Version: v,
				Data:    value.B,
			})
		}
	}

	return records
}

func isKey(k string) bool {
	return strings.HasPrefix(k, keyPrefix)
}

func versionFromKey(key string) int {
	v, _ := strconv.Atoi(key[len(keyPrefix):])
	return v
}

func keyFromVersion(version int) string {
	return fmt.Sprintf("%s%d", keyPrefix, version)
}

func filterRecords(start, end int, records []eventing.Record) []eventing.Record {

	filtered := []eventing.Record{}

	for _, record := range records {
		if record.Version >= start && record.Version <= end {
			filtered = append(filtered, record)
		}
	}

	return filtered
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

func (s *store) Save(ctx context.Context, id string, records []eventing.Record) error {
	if len(records) == 0 {
		return nil
	}

	return nil
}

func (s *store) makeUpdateInput(id string, records []eventing.Record) (*d.UpdateItemInput, error) {
	partition := selectPartition(records[0].Version, s.itemsPerPartition)

	conditions := []expression.ConditionBuilder{}
	update := expression.Add(expression.Name("revision"), expression.Value(1))
	for _, record := range records {
		key := keyFromVersion(record.Version)
		conditions = append(conditions, expression.AttributeNotExists(expression.Name(key)))
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
