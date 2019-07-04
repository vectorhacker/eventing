package eventing

import (
	"context"
	"errors"
)

// Version constants
const (
	EndOfStream = 0
)

// Store represents an event store for serialized events
// this is used for supporting different backends
type Store interface {
	Load(context.Context, string, int, int) ([]Record, error)
	Save(context.Context, string, []Record) error
}

type memoryStore map[string][]Record

func (m memoryStore) Load(_ context.Context, id string, start, end int) ([]Record, error) {
	records, ok := m[id]
	if !ok {
		return nil, errors.New("not in store")
	}

	records = records[start+1:]
	if end > 0 {
		records = records[:end]
	}
	return records, nil
}

func (m memoryStore) Save(_ context.Context, id string, records []Record) error {
	m[id] = append(m[id], records...)
	return nil
}
