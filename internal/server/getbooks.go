package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/log"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"github.com/google/uuid"
)

// GetApiV1Books handles the API request to search for books based on a query string.
// It calls the OpenLibrary service to fetch book details and formats the response accordingly.
//
// Parameters:
// - ctx: The context for the request, allowing for cancellation and timeout control.
// - request: The request object containing query parameters for book search.
//
// Returns:
// - betterreads.GetApiV1BooksResponseObject: The API response containing book details or an error message.
// - error: Always nil, as errors are encapsulated in the response object. In OpenAPI, the error is reserved for internal server errors.
//
//nolint:revive // must satisfy interface
func (s *Server) GetApiV1Books(ctx context.Context, request betterreads.GetApiV1BooksRequestObject) (betterreads.GetApiV1BooksResponseObject, error) {
	if err := verifyGetAPIV1BooksRequest(request); err != nil {
		//nolint:nilerr // errors are reserved for internal errors
		return betterreads.GetApiV1Books400JSONResponse{
			Code: "BAD_REQUEST",
			Details: &map[string]any{
				"error":        err.Error(),
				"reference_id": uuid.New(),
			},
			Message: "search books - invalid request",
		}, nil
	}

	searchResult, err := s.OpenLibrary.SearchBooks(
		ctx,
		getStringFromPointer(request.Params.Query),
		request.Params.Title,
		request.Params.Author,
		request.Params.Subject,
	)
	if err != nil {
		var resp betterreads.GetApiV1BooksResponseObject
		switch {
		case errors.Is(err, openlibrary.ErrBadRequest):
			resp = betterreads.GetApiV1Books400JSONResponse{
				Code: "BAD_REQUEST",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - bad request",
			}
		case errors.Is(err, openlibrary.ErrNotFound):
			resp = betterreads.GetApiV1Books404JSONResponse{
				Code: "NOT_FOUND",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - not found",
			}
		case errors.Is(err, openlibrary.ErrInternalServer):
			resp = betterreads.GetApiV1Books500JSONResponse{
				Code: "INTERNAL_SERVER_ERROR",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - internal server error",
			}
		default:
			log.Warn("unhandled error", map[string]error{"error": err})
			resp = betterreads.GetApiV1Books500JSONResponse{
				Code: "UNKNOWN",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - unknown",
			}
		}
		return resp, nil
	}

	books := make([]betterreads.Book, 0)
	for _, book := range searchResult.Books {
		books = append(books, betterreads.Book{
			Id:            book.CoverEditionKey,
			Title:         book.Title,
			AuthorName:    book.AuthorName,
			AuthorId:      book.AuthorKey,
			BookImage:     book.CoverImage,
			PublishedYear: book.PublishYear,
			Isbn:          book.ISBN,
			RatingCount:   book.RatingCount,
			RatingAverage: book.RatingAverage,
			Source:        book.Source,
		})
	}
	return betterreads.GetApiV1Books200JSONResponse{Books: books}, nil
}

func verifyGetAPIV1BooksRequest(request betterreads.GetApiV1BooksRequestObject) error {
	if (request.Params.Query == nil || getStringFromPointer(request.Params.Query) == "") &&
		request.Params.Title == nil &&
		request.Params.Author == nil &&
		request.Params.Subject == nil {
		return errors.New("must pass one search parameter")
	}
	return nil
}
