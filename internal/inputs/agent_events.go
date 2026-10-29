package inputs

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Interactive-AI-Labs/interactive-cli/internal/clients/deployment"
)

// MessageEvent is the body that sends a customer message.
func MessageEvent(message string) json.RawMessage {
	body, _ := json.Marshal(map[string]string{
		"kind":    deployment.EventKindMessage,
		"source":  deployment.EventSourceCustomer,
		"message": message,
	})
	return body
}

// BuildEventBody is the event to send: a customer message, or JSON from a file or stdin (-).
func BuildEventBody(message, file string, stdin io.Reader) (json.RawMessage, error) {
	if message != "" {
		return MessageEvent(message), nil
	}
	if file == "" {
		return nil, fmt.Errorf("--message must not be empty")
	}
	var raw []byte
	var err error
	if file == "-" {
		raw, err = io.ReadAll(stdin)
	} else {
		raw, err = os.ReadFile(file)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read event file: %w", err)
	}
	var event map[string]json.RawMessage
	if err := json.Unmarshal(raw, &event); err != nil || event == nil {
		return nil, fmt.Errorf("event file must be a JSON object")
	}
	return raw, nil
}
