package main

import (
	"go.temporal.io/api/history/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

// sanitizeEvent removes all Payloads from the given history event's attributes. This helps mitigate the impact of
// large workflow histories (temporal permits up to 50mb) on small LLM context windows (~2mb). This is just best
// effort - it assumes that largeness is caused by the payloads.
func sanitizeEvent(event *history.HistoryEvent) {
	sanitizeRecursively(event.ProtoReflect())
}

var REPLACEMENT_VALUE = protoreflect.ValueOf(nil)

// HistoryEvents are highly polymorphic (today: 54 different types), and Temporal could add new types at any time (most
// recent time: launching Nexus). Let's sanitize via convention, rather than a hard-coded list of history event types
// and their structure.
func sanitizeRecursively(m protoreflect.Message) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList():
			// Avoid lists of non-messages
			if fd.Kind() != protoreflect.MessageKind {
				return true
			}

			list := v.List()
			for i := 0; i < list.Len(); i++ {
				item := list.Get(i).Message()
				if isPayload(item) {
					// Proto lists are homogeneous - if any items are payloads, all items are payloads
					list.Truncate(0)
				} else {
					sanitizeRecursively(item)
				}
			}
		case fd.IsMap():
			// Avoid maps of non-messages
			if fd.MapValue().Kind() != protoreflect.MessageKind {
				return true
			}

			mapp := v.Map()
			mapp.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
				val := v.Message()
				if isPayload(val) {
					mapp.Clear(k)
				} else {
					sanitizeRecursively(val)
				}

				return true
			})
		default:
			if fd.Kind() == protoreflect.MessageKind {
				msg := v.Message()
				if isPayload(msg) {
					m.Clear(fd)
				} else {
					sanitizeRecursively(msg)
				}
			}
		}

		return true
	})
}

func isPayload(m protoreflect.Message) bool {
	fullType := string(m.Descriptor().FullName())
	return strings.HasSuffix(fullType, ".Payload") || strings.HasSuffix(fullType, ".Payloads")
}
