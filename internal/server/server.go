package server

import (
	"net/http"

	"github.com/celestialdragonfly/betterreads-platform/internal/contracts"
	"github.com/celestialdragonfly/betterreads-platform/internal/data"
	"github.com/celestialdragonfly/betterreads-platform/internal/dependency/google"
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
	config        *Config
	logger        *log.Logger
	DB            *data.SQL
	Handler       http.Handler
	GoogleBookAPI *google.BookAPI
}

func NewBetterReads(l *log.Logger, database *data.SQL, bookAPI *google.BookAPI, cfg *Config) *BetterReads {
	br := &BetterReads{
		logger:        l,
		DB:            database,
		GoogleBookAPI: bookAPI,
		config:        cfg,
	}
	br.Handler = routes(br)
	return br
}

func routes(br *BetterReads) http.Handler {
	router := httprouter.New()

	// Health Check
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", br.HealthcheckHandler)

	// User Management
	router.HandlerFunc(http.MethodPost, "/api/v1/users", br.CreateUserHandler)
	router.HandlerFunc(http.MethodGet, "/api/v1/users/:userid", br.GetUserHandler)

	// Book Management
	router.HandlerFunc(http.MethodGet, "/api/v1/books/:query", br.GetBooks)

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

func (br *BetterReads) GetBooks(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())

	br.GoogleBookAPI.GetBooks(params.ByName("query"))
}

// HEALTHCHECK-API
// HealthCheckHandler
func (br *BetterReads) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := contracts.HealthCheckResponse{
		Status:      "available",
		Environment: br.config.Env,
		Version:     br.config.Version,
	}

	err := json.Marshal(w, http.StatusOK, resp, nil)
	if err != nil {
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}

func (br *BetterReads) unimplemented(w http.ResponseWriter, r *http.Request) {
	err := json.Marshal(w, http.StatusNotImplemented, struct{ Message string }{Message: "method not implemented"}, nil)
	if err != nil {
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
