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
	router.HandlerFunc(http.MethodGet, "/api/v1/books/search", br.GetGoogleBooks)                              // Query google books api for all books matching search param.
	router.HandlerFunc(http.MethodGet, "/api/v1/books/library", br.GetBooksForUser)                            // Get all books from a user's library.
	router.HandlerFunc(http.MethodGet, "/api/v1/books/library/:bookid", br.GetBookForUser)                     // Retrieves a signle book the user's library given the book id.
	router.HandlerFunc(http.MethodDelete, "/api/v1/books/library/:bookid", br.RemoveBookForUser)               // Remove a book from a user's library given the book id.
	router.HandlerFunc(http.MethodPost, "/api/v1/books/library/:bookid/read", br.AddBookToReadForUser)         // Add book to read list.
	router.HandlerFunc(http.MethodPost, "/api/v1/books/library/:bookid/reading", br.AddBookToReadingForUser)   // Add book to reading list.
	router.HandlerFunc(http.MethodPost, "/api/v1/books/library/:bookid/wishlist", br.AddBookToWishlistForUser) // Add book to want to read list.

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

// BOOKS-API
// GetGoogleBooks query google books api for all books matching search param.
func (br *BetterReads) GetGoogleBooks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	wildcard := query.Get("q")

	booksResponse, err := br.GoogleBookAPI.GetBooks(br.logger, wildcard)
	if err != nil {
		json.Response(w, http.StatusInternalServerError, contracts.NewErrorResponse(err.Error()), nil)
	}

	books := make([]contracts.GoogleBook, 0)
	for _, item := range booksResponse.Items {
		book := contracts.GoogleBook{
			ID:             item.ID,
			Etag:           item.Etag,
			SelfLink:       item.SelfLink,
			Title:          item.VolumeInfo.Title,
			Subtitle:       item.VolumeInfo.Subtitle,
			Authors:        item.VolumeInfo.Authors,
			Categories:     item.VolumeInfo.Categories,
			AverageRating:  item.VolumeInfo.AverageRating,
			RatingsCount:   item.VolumeInfo.RatingsCount,
			MaturityRating: item.VolumeInfo.MaturityRating,
			SmallThumbnail: item.VolumeInfo.ImageLinks.SmallThumbnail,
			Thumbnail:      item.VolumeInfo.ImageLinks.Thumbnail,
			PreviewLink:    item.VolumeInfo.PreviewLink,
			InfoLink:       item.VolumeInfo.InfoLink,
		}
		books = append(books, book)
	}

	json.Response(w, http.StatusOK, contracts.GetBooksResponse{Books: books}, nil)
}

// GetBooksForUser retrieves all books from a user's library.
func (br *BetterReads) GetBooksForUser(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)

	json.Response(w, http.StatusOK, contracts.GetBooksForUserResponse{}, nil)
}

// GetBookForUser retrieves a signle book the user's library given the book id.
func (br *BetterReads) GetBookForUser(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// RemoveBookForUser removes a book from a user's library given the book id.
func (br *BetterReads) RemoveBookForUser(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// AddBookToReadForUser updates the category type for a book in a user's library to "read".
func (br *BetterReads) AddBookToReadForUser(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// AddBookToReadingForUser updates the category type for a book in a user's library to "reading".
func (br *BetterReads) AddBookToReadingForUser(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// AddBookToWishlistForUser updates the category type for a book in a user's library to "want to read/wishlist".
func (br *BetterReads) AddBookToWishlistForUser(w http.ResponseWriter, r *http.Request) {
	br.unimplemented(w, r)
}

// HEALTHCHECK-API
// HealthCheckHandler
func (br *BetterReads) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	resp := contracts.HealthCheckResponse{
		Status:      "available",
		Environment: br.config.Env,
		Version:     br.config.Version,
	}

	json.Response(w, http.StatusOK, resp, nil)

}

func (br *BetterReads) unimplemented(w http.ResponseWriter, r *http.Request) {
	json.Response(w, http.StatusNotImplemented, struct{ Message string }{Message: "method not implemented"}, nil)
}
