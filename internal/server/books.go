package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/log"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"github.com/google/uuid"
)

// (GET /api/v1/books).
func (s *Server) SearchBooks(ctx context.Context, request betterreads.SearchBooksRequestObject) (betterreads.SearchBooksResponseObject, error) {
	if err := verifyGetAPIV1BooksRequest(request); err != nil {
		//nolint:nilerr // errors are reserved for internal errors
		return betterreads.SearchBooks400JSONResponse{
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
		var resp betterreads.SearchBooksResponseObject
		switch {
		case errors.Is(err, openlibrary.ErrBadRequest):
			resp = betterreads.SearchBooks400JSONResponse{
				Code: "BAD_REQUEST",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - bad request",
			}
		case errors.Is(err, openlibrary.ErrNotFound):
			resp = betterreads.SearchBooks400JSONResponse{ // todo add 404 back
				Code: "NOT_FOUND",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - not found",
			}
		case errors.Is(err, openlibrary.ErrInternalServer):
			resp = betterreads.SearchBooks500JSONResponse{
				Code: "INTERNAL_SERVER_ERROR",
				Details: &map[string]any{
					"error":        err.Error(),
					"reference_id": uuid.New(),
				},
				Message: "search books - internal server error",
			}
		default:
			log.Warn("unhandled error", map[string]error{"error": err})
			resp = betterreads.SearchBooks500JSONResponse{
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
			Source:        betterreads.BookSourceOpenLibrary, // todo add mapper
		})
	}
	return betterreads.SearchBooks200JSONResponse{
		Body: betterreads.GetBooksResponse{
			Books: books,
		},
	}, nil
}

func verifyGetAPIV1BooksRequest(request betterreads.SearchBooksRequestObject) error {
	if (request.Params.Query == nil || getStringFromPointer(request.Params.Query) == "") &&
		request.Params.Title == nil &&
		request.Params.Author == nil &&
		request.Params.Subject == nil {
		return errors.New("must pass one search parameter")
	}
	return nil
}
