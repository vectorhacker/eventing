package test

import (
	"testing"

	"github.com/gogo/protobuf/proto"
)

func TestTest(t *testing.T) {
	b := NewBuilder("123", 0)

	b.EventA(KnownLanguages_ENGLISH, &EventA_InnerMessage{
		Hello: "World",
	})

	b.EventB(&OuterMessage{
		Hola: "Mundo",
	})

	b.EventB(nil)

	if len(b.Events) != 3 {
		t.Fatal("b.Events should have 3 events: ", b.Events)
	}

	s := NewSerializer()

	event := b.Events[2]

	record, err := s.Serialize(event)
	if err != nil {
		t.Fatal("ERROR:", err)
	}

	got, err := s.Deserialize(record)
	if err != nil {
		t.Fatal("ERROR:", err)
	}

	if !proto.Equal(event.(*EventB), got.(*EventB)) {
		t.Fatalf("Not equal:\nexpected: %##v\n actual: %##v", event, got)
	}
}
