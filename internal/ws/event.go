package ws

import "encoding/json"

const (
	EventNotificationCreated = "notification.created"
)

type EventMessage[T any] struct {
	Event string `json:"event"`
	Data  T      `json:"data"`
}

func MarshalEvent[T any](event string, data T) ([]byte, error) {
	return json.Marshal(EventMessage[T]{
		Event: event,
		Data:  data,
	})
}
