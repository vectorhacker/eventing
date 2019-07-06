package test

import (
	"testing"
)

func TestTest(t *testing.T) {
	b := NewBuilder("123", 0)

	b.EventA(KnownLanguages_ENGLISH, &EventA_InnerMessage{
		Hello: "World",
	})

	b.EventB(&OuterMessage{
		Hola: "Mundo",
	})

	if len(b.Events) != 2 {
		t.Fatal("b.Events should have 2 events: ", b.Events)
	}

	t.Logf("%v", b.Events)
}
