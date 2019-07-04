package eventing

import (
	"reflect"
)

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
