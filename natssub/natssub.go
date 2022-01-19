package natssub

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Key string
}

func Parse(message []byte) (string, error) {
	event := Event{}
	if err := json.Unmarshal(message, &event); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON; %w", err)
	}
	return event.Key, nil
}
