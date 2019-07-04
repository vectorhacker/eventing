package eventing

import (
	"encoding/json"
	"reflect"
)

// EventNamer is an event that can name itself
type EventNamer interface {
	EventName() string
}

// JSONSerializer implements the Serializer interface
type JSONSerializer struct {
	types map[string]reflect.Type
}

type jsonEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// NewJSONSerializer serializes events into/out of JSON,
// it uses reflection to get the new events and so events should
// not have any behaviour depending on some initialization like a custom
// constructor
func NewJSONSerializer(events ...Event) *JSONSerializer {
	types := make(map[string]reflect.Type, len(events))

	for _, event := range events {
		types[eventName(event)] = typeOf(event)
	}

	return &JSONSerializer{
		types: types,
	}
}

// Serialize implements the Serializer interface
func (s *JSONSerializer) Serialize(event Event) (Record, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return Record{}, err
	}

	data, err := json.Marshal(jsonEvent{
		Type:    eventName(event),
		Payload: payload,
	})
	if err != nil {
		return Record{}, err
	}

	return Record{
		Version: event.EventVersion(),
		Data:    data,
	}, err
}

// Deserialize implements the Serializer interface
func (s *JSONSerializer) Deserialize(record Record) (Event, error) {
	j := &jsonEvent{}
	if err := json.Unmarshal(record.Data, j); err != nil {
		return nil, err
	}

	e := s.new(j.Type)

	if err := json.Unmarshal(j.Payload, e); err != nil {
		return nil, err
	}

	return e, nil
}

func (s *JSONSerializer) new(eventType string) Event {
	t := s.types[eventType]
	v := reflect.New(t)

	return v.Interface().(Event)
}

func typeOf(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func eventName(event Event) string {
	if namer, ok := event.(EventNamer); ok {
		return namer.EventName()
	}

	t := typeOf(event)
	return t.Name()
}
