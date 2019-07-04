package dynamodb

import (
	"fmt"
	"strconv"
	"strings"

	d "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/vectorhacker/eventing"
)

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
