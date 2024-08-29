package json

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Response is helper for sending responses.
func Response[T any](w http.ResponseWriter, status int, data T, headers http.Header) {
	// Encode the data to JSON, returning the error if there was one.
	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "The server encountered a problem and could not process your request. S=00001", http.StatusInternalServerError)
	}

	// appending a new line to make this easier to read via a terminal.
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(js); err != nil {
		http.Error(w, "The server encountered a problem and could not process your request. S=00002", http.StatusInternalServerError)
	}
}

// Limit the size of the request body to 1MB.
const maxBytes = 1_048_576

var (
	syntaxError           *json.SyntaxError
	unmarshalTypeError    *json.UnmarshalTypeError
	invalidUnmarshalError *json.InvalidUnmarshalError
	maxBytesError         *http.MaxBytesError
)

// TODO abstract away returning errors from the client in favor of an error type return with a reference_id. Log the current errors.

// Request is helper for reading requests.
func Request(w http.ResponseWriter, r *http.Request, dst any) error {

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			return fmt.Errorf("body contained invalid json. error: %w", err)

		default:
			return fmt.Errorf("unhandled json unmarshal error. error: %w", err)
		}
	}

	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
