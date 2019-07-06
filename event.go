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

// EventModel implements the Event interface
type EventModel struct {
	ID string
	Version int
	At time.Time
}


// AggregateID implements the Event interface
func (e EventModel) AggregateID() string {
	return e.ID
}


// EventVersion implements the Event interface
func (e EventModel) EventVersion() int {
	return e.Version
}

// EventAt implements the Event interface
func (e EventModel) EventAt() time.Time {
	return e.At
}

// Record is an event that was serialzied
type Record struct {
	Data    []byte
	Version int
}

// SortRecords sorts records using their version
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
