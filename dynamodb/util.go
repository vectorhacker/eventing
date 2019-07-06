package dynamodb

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	d "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/vectorhacker/eventing"
)

func DynamoDB(region, endpoint string) (*dynamodb.DynamoDB, error) {
	cfg := &aws.Config{
		Region: aws.String(region),
	}
	if endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}

	s, err := session.NewSession(cfg)
	if err != nil {
		return nil, err
	}
	return dynamodb.New(s), nil
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
		if record.Version < start {
			continue
		}

		if end > 0 && record.Version > end {
			continue
		}

		filtered = append(filtered, record)
	}

	return filtered
}
