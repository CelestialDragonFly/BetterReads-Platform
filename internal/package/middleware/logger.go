package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	iErrors "github.com/celestialdragonfly/betterreads-platform/internal/package/errors"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type responseLogger struct {
	http.ResponseWriter
	statusCode int
	buf        *bytes.Buffer
}

func (rl *responseLogger) Write(p []byte) (int, error) {
	rl.buf.Write(p)
	return rl.ResponseWriter.Write(p)
}

func (rl *responseLogger) WriteHeader(statusCode int) {
	rl.statusCode = statusCode
	rl.ResponseWriter.WriteHeader(statusCode)
}

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Intercept response
		logWriter := &responseLogger{ResponseWriter: w, buf: &bytes.Buffer{}}
		next(logWriter, r)

		// success
		if 299 >= logWriter.statusCode && logWriter.statusCode >= 200 {
			logger.Info("handled request",
				"user-id", logWriter.Header().Get("user-id"),
				"request", r.URL.Path,
				"status_code", logWriter.statusCode,
			)
		} else {
			var iErr iErrors.Error
			err := json.NewDecoder(logWriter.buf).Decode(&iErr)
			if err != nil {
				logger.Error("encountered error",
					"user-id", logWriter.Header().Get("user-id"),
					"request", r.URL.Path,
					"status_code", logWriter.statusCode,
					"err", logWriter.buf.String(),
				)
				return
			}
			logger.Error("encountered error",
				"user-id", logWriter.Header().Get("user-id"),
				"request", r.URL.Path,
				"status_code", logWriter.statusCode,
				"err", iErr.Message,
				"reference_id", iErr.ReferenceID,
			)
		}
	}
}
