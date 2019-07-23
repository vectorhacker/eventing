package eventing

import (
	"context"
	"testing"
	"time"
)

type testCommand struct {
	CommandModel
}

type testEvent struct {
	EventModel
}

type testAggregate struct {
	ID      string
	Version int
}

func (t *testAggregate) On(event Event) error {

	switch e := event.(type) {
	case *testEvent:
		t.ID = e.ID
		t.Version = e.Version
	}

	return nil
}

func (t *testAggregate) Apply(_ context.Context, cmd Command) ([]Event, error) {
	return []Event{
		&testEvent{
			EventModel: EventModel{ID: cmd.AggregateID(), Version: t.Version + 1, At: time.Now()},
		},
	}, nil
}

func TestRepository(t *testing.T) {

	r := New(&testAggregate{}, WithSerializer(NewJSONSerializer(&testEvent{})))

	version, err := r.Dispatch(context.Background(), testCommand{CommandModel: CommandModel{ID: "123"}})
	if err != nil {
		t.Fatal("ERROR:", err)
	}

	if version != 1 {
		t.Fatal("VERSION SHOULD BE 1", version)
	}

	a, err := r.Load(context.Background(), "123")
	if err != nil {
		t.Fatal("ERROR:", err)
	}

	testA, ok := a.(*testAggregate)
	if !ok {
		t.Fatal("should be testAggregate")
	}

	if testA.ID != "123" {
		t.Fatal("ID should be 123 got:", testA.ID)
	}
}
