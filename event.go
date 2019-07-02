package eventing

import (
	"sort"
	"time"
)

// Event interface
type Event interface {
	AggregateID() string
	EventVersion() int
	EventAt() time.Time
}

// Record is an event that was serialzied
type Record struct {
	Data    []byte
	Version int
}

// Sort sorts records
func SortRecords(r []Record) {
	sort.Slice(r, func(i, j int) bool {
		return r[i].Version < r[j].Version
	})
}

// Serializer serializes/deserializes an event
type Serializer interface {
	Serialize(event Event) (Record, error)
	Deserialize(record Record) (Event, error)
}
