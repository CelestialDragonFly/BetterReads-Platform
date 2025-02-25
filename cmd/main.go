package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/server"
	strictnethttp "github.com/oapi-codegen/runtime/strictmiddleware/nethttp"
)

var (
	Host = GetDefault("BR_HOST", "localhost")
	Port = GetDefault("BR_PORT", "8080")
)

func main() {
	server := server.NewServer(&server.Config{})

	strictHandler := betterreads.NewStrictHandler(
		server,
		[]betterreads.StrictMiddlewareFunc{
			loggingMiddleware(),
		},
	)
	httpHandler := betterreads.Handler(strictHandler)

	srv := &http.Server{
		Addr:    ":" + Port,
		Handler: httpHandler,
	}

	log.Printf("Server starting on port %s", Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// GetDefault returns the string value of the environment variable, or a default
// value if the environment variable is not defined or is an empty string
func GetDefault(envVar, defaultValue string) string {
	if v, ok := os.LookupEnv(envVar); ok && len(v) > 0 {
		return v
	}
	return defaultValue
}

// GetBoolDefault returns the boolean value of the environment variable, or a default
// value if the environment variable is not defined or is an empty string
func GetBoolDefault(envVar string, defaultValue bool) bool {
	val := GetDefault(envVar, strconv.FormatBool(defaultValue))
	if b, err := strconv.ParseBool(val); err == nil {
		return b
	}
	return defaultValue
}

// GetIntDefault returns the int value of the environment variable, or a default
// value if the environment variable is not defined or is an empty string
func GetIntDefault(envVar string, defaultValue int) int {
	val := GetDefault(envVar, strconv.Itoa(defaultValue))
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	return defaultValue
}

// GetInt64Default returns the int64 value of the environment variable, or a default
// value if the environment variable is not defined or is an empty string
func GetInt64Default(envVar string, defaultValue int64) int64 {
	val := GetDefault(envVar, strconv.FormatInt(defaultValue, 16))
	if i, err := strconv.ParseInt(val, 10, 64); err == nil {
		return i
	}
	return defaultValue
}

// GetFloatDefault returns the float64 value of the environment variable, or a default
// value if the environment variable is not defined or is an empty string
func GetFloatDefault(envVar string, defaultValue float64) float64 {
	val := GetDefault(envVar, strconv.FormatFloat(defaultValue, 'E', -1, 64))
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}
	return defaultValue
}

// GetDurationDefault returns the time.Duration value of the environment variable, or a default
// value if the environment variable is not defined or is an empty string
func GetDurationDefault(envVar string, defaultValue time.Duration) time.Duration {
	val := GetDefault(envVar, defaultValue.String())
	if t, err := time.ParseDuration(val); err == nil {
		return t
	}
	return defaultValue
}

// loggingMiddleware creates a StrictMiddlewareFunc for logging
func loggingMiddleware() betterreads.StrictMiddlewareFunc {
	return func(f strictnethttp.StrictHTTPHandlerFunc, operationID string) strictnethttp.StrictHTTPHandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (response interface{}, err error) {
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

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
