package server

import (
	"net/http"

	"github.com/celestialdragonfly/betterreads-platform/internal/contract"
	"github.com/celestialdragonfly/betterreads-platform/internal/package/json"
	"github.com/celestialdragonfly/betterreads-platform/internal/package/log"
	"github.com/julienschmidt/httprouter"
)

type Config struct {
	Port    int
	Env     string
	Version string
}

type BetterReads struct {
	config  *Config
	logger  *log.Logger
	Handler http.Handler
}

func NewBetterReads(l *log.Logger, cfg *Config) *BetterReads {
	br := &BetterReads{
		logger: l,
		config: cfg,
	}
	br.Handler = routes(br)
	return br
}

func routes(br *BetterReads) http.Handler {
	// Initialize a new httprouter router instance.
	router := httprouter.New()

	// Health Check
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", br.HealthcheckHandler)

	// User Management
	router.HandlerFunc(http.MethodPost, "/api/v1/users", br.CreateUserHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:userid", br.GetUserHandler)

	return router
}

// USERS-API
// CreateUserHandler creates a new BetterReads user.
func (br *BetterReads) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// GetUserHandler retrieves a new BetterReads user.
func (br *BetterReads) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// HEALTHCHECK-API
// HealthCheckHandler
func (br *BetterReads) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := contract.HealthCheckResponse{
		Status:      "available",
		Environment: br.config.Env,
		Version:     br.config.Version,
	}

	err := json.WriteJSON(w, http.StatusOK, resp, nil)
	if err != nil {
		// TODO remove this log
		br.logger.Error(err.Error(), nil)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

func (br *BetterReads) unimplemented(w http.ResponseWriter, r *http.Request) {
	err := json.WriteJSON(w, http.StatusNotImplemented, struct{ Message string }{Message: "method not implemented"}, nil)
	if err != nil {
		// TODO remove this log
		br.logger.Error(err.Error(), nil)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
