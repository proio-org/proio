package proio

import (
	"bytes"
	"testing"
)

func TestPushUpdate1(t *testing.T) {
	buffer := &bytes.Buffer{}

	writer := NewWriter(buffer)
	writer.PushMetadata("key1", []byte("value1"))
	writer.PushMetadata("key2", []byte("value2"))
	event := NewEvent()
	writer.Push(event)
	writer.PushMetadata("key2", []byte("value3"))
	writer.Push(event)
	writer.PushMetadata("key1", []byte("value4"))
	writer.PushMetadata("key2", []byte("value5"))
	writer.Push(event)
	writer.Close()

	reader := NewReader(buffer)
	event1, _ := reader.Next()
	event2, _ := reader.Next()
	event3, _ := reader.Next()
	if string(event1.Metadata["key1"]) != "value1" {
		t.Errorf("%v -> %v instead of %v", "key1", event1.Metadata["key1"], "value1")
	}
	if string(event1.Metadata["key2"]) != "value2" {
		t.Errorf("%v -> %v instead of %v", "key2", event1.Metadata["key2"], "value2")
	}
	if string(event2.Metadata["key2"]) != "value3" {
		t.Errorf("%v -> %v instead of %v", "key2", event2.Metadata["key2"], "value3")
	}
	if string(event3.Metadata["key1"]) != "value4" {
		t.Errorf("%v -> %v instead of %v", "key1", event3.Metadata["key1"], "value4")
	}
	if string(event3.Metadata["key2"]) != "value5" {
		t.Errorf("%v -> %v instead of %v", "key2", event3.Metadata["key2"], "value5")
	}
	reader.Close()

	writer = NewWriter(buffer)
	writer.Push(event1)
	writer.Push(event2)
	writer.Push(event3)
	writer.Close()

	reader = NewReader(buffer)
	event1, _ = reader.Next()
	event2, _ = reader.Next()
	event3, _ = reader.Next()
	if string(event1.Metadata["key1"]) != "value1" {
		t.Errorf("%v -> %v instead of %v", "key1", event1.Metadata["key1"], "value1")
	}
	if string(event1.Metadata["key2"]) != "value2" {
		t.Errorf("%v -> %v instead of %v", "key2", event1.Metadata["key2"], "value2")
	}
	if string(event2.Metadata["key2"]) != "value3" {
		t.Errorf("%v -> %v instead of %v", "key2", event2.Metadata["key2"], "value3")
	}
	if string(event3.Metadata["key1"]) != "value4" {
		t.Errorf("%v -> %v instead of %v", "key1", event3.Metadata["key1"], "value4")
	}
	if string(event3.Metadata["key2"]) != "value5" {
		t.Errorf("%v -> %v instead of %v", "key2", event3.Metadata["key2"], "value5")
	}
	reader.Close()
}
