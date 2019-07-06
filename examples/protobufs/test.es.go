package test

import (
	"errors"
	proto "github.com/gogo/protobuf/proto"
	eventing "github.com/vectorhacker/eventing"
	"time"
)

type Builder struct {
	ID      string
	Events  []eventing.Event
	Version int
}

func (b *Builder) nextVersion() {
	b.Version++
}

func NewBuilder(id string, version int) *Builder {
	return &Builder{
		ID:      id,
		Version: version,
	}
}

type Serializer struct{}

func NewSerializer() eventing.Serializer {
	return &Serializer{}
}

func (s Serializer) Serialize(event eventing.Event) (eventing.Record, error) {
	container := &TestEvent{}
	switch e := event.(type) {
	case *EventA:
		container.Event = &TestEvent_EventA{EventA: e}
	case *EventB:
		container.Event = &TestEvent_EventB{EventB: e}
	}

	data, err := proto.Marshal(container)

	if err != nil {
		return eventing.Record{}, err
	}

	return eventing.Record{
		Data:    data,
		Version: event.EventVersion(),
	}, nil
}

func (s Serializer) Deserialize(record eventing.Record) (eventing.Event, error) {
	container := &TestEvent{}
	err := proto.Unmarshal(record.Data, container)
	if err != nil {
		return nil, err
	}
	switch container.Event.(type) {
	case *TestEvent_EventA:
		return container.GetEventA(), nil
	case *TestEvent_EventB:
		return container.GetEventB(), nil
	}

	return nil, errors.New("No event")
}

// AggregateID implements the Event interface
func (e *EventA) AggregateID() string {
	return e.Id
}

// EventVersion implements the Event interface
func (e *EventA) EventVersion() int {
	return int(e.Version)
}

// EventAt implements the Event interface
func (e *EventA) EventAt() time.Time {
	return time.Unix(e.At, 0)
}

// EventName implements the EventNamer interface
func (EventA) EventName() string {
	return "EventA"
}

func (b *Builder) EventA(language KnownLanguages, message *EventA_InnerMessage) {
	e := &EventA{
		At:       time.Now().Unix(),
		Id:       b.ID,
		Language: language,
		Message:  message,
		Version:  int32(b.Version),
	}
	b.Events = append(b.Events, e)
	b.nextVersion()
}

// AggregateID implements the Event interface
func (e *EventB) AggregateID() string {
	return e.Id
}

// EventVersion implements the Event interface
func (e *EventB) EventVersion() int {
	return int(e.Version)
}

// EventAt implements the Event interface
func (e *EventB) EventAt() time.Time {
	return time.Unix(e.At, 0)
}

// EventName implements the EventNamer interface
func (EventB) EventName() string {
	return "EventB"
}

func (b *Builder) EventB(message *OuterMessage) {
	e := &EventB{
		At:      time.Now().Unix(),
		Id:      b.ID,
		Message: message,
		Version: int32(b.Version),
	}
	b.Events = append(b.Events, e)
	b.nextVersion()
}
