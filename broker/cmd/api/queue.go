package main

import (
	"broker/internal/event"
	"encoding/json"
)

type EventPayload struct {
	Name string `json:"name"`
	Data any    `json:"data"`
}

func (c *Config) pushToQueue(name string, topic string, severity string, data any) error {
	emitter, err := event.NewEventEmitter(c.Rabbit)
	if err != nil {
		return err
	}

	payload := EventPayload{
		Name: name,
		Data: data,
	}

	j, _ := json.Marshal(&payload)

	err = emitter.Push(string(j), severity, topic)
	if err != nil {
		return err
	}

	return nil
}
