package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Error struct {
	Message     string `json:"msg"`
	ReferenceID string `json:"reference_id,omitempty"`
}

func NewHttpError(w *http.ResponseWriter, message string, err error) {
	json.NewEncoder(*w).Encode(Error{
		Message:     fmt.Sprintf("%s: %v", message, err),
		ReferenceID: uuid.NewString(),
	})
}

func WrapError(error1, error2 error) error {
	return fmt.Errorf("%w: %w", error1, error2)
}
