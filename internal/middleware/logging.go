package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

// responseWriter is a custom wrapper around the http.ResponseWriter to capture
// the status code and body of the response for logging purposes.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// WriteHeader overrides the default WriteHeader method to capture the status code
// for logging purposes before passing it to the original ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code // Capture em like Pokemon.
	rw.ResponseWriter.WriteHeader(code)
}

// Write overrides the default Write method to capture the response body
// into the 'body' field of the responseWriter, enabling logging of the response
// body if the status is non-2xx.
func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = body // Capture em like Pokemon.
	return rw.ResponseWriter.Write(body)
}

// Logging returns a middleware function that logs the details of the HTTP request
// and response. It captures the request's path, method, status code, duration,
// and response body (if non-2xx) and logs this information using the slog package.
func Logging() betterreads.StrictMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, _ string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default if not set
			}

			start := time.Now()
			resp, err := f(ctx, wrapped, r, request)
			duration := time.Since(start)

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			level := slog.LevelInfo

			// Capture error if status is non-2xx
			errToLog := err
			if wrapped.statusCode < 200 || wrapped.statusCode > 299 {
				level = slog.LevelError
				// If the error is nil, assign the response body as the error
				if err == nil {
					errToLog = errors.New(string(wrapped.body))
				}
			}

			logger.LogAttrs(
				ctx,
				level,
				"handled request",
				slog.String("path", r.URL.Path),
				slog.String("method", r.Method),
				slog.Int("status", wrapped.statusCode),
				slog.String("duration", duration.String()),
				slog.Any("error", errToLog),
			)

			return resp, err
		}
	}
}
