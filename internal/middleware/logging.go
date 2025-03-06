package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging() betterreads.StrictMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, _ string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
			// Wrap response writer to capture status
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default if not set
			}

			// Call the handler with wrapped writer
			resp, err := f(ctx, wrapped, r, request)

			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
			level := slog.LevelInfo
			if err != nil {
				level = slog.LevelError
			}
			logger.LogAttrs(
				context.TODO(),
				level,
				"Handled request",
				slog.String("Path", r.URL.Path),
				slog.String("Method", r.Method),
				slog.Int("Status", wrapped.statusCode),
				slog.Any("Error", err),
			)

			return resp, err
		}
	}
}
