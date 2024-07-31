package json

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteJSON() is helper for sending responses.
func WriteJSON[T any](w http.ResponseWriter, status int, data T, headers http.Header) error {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// appending a new line to make this easier to read via a terminal.
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(js); err != nil {
		return fmt.Errorf("error writing json: %w", err)
	}

	return nil
}

func ReadJSON() {}
