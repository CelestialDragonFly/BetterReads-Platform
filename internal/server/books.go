package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// (GET /api/v1/books).
// SearchBooks implements betterreads.BetterReadsServiceServer.
func (s *Server) SearchBooks(ctx context.Context, request *betterreads.SearchBooksRequest) (*betterreads.SearchBooksResponse, error) {
	if err := verifySearchBooksRequest(request); err != nil {
		return nil, err
	}

	searchResult, err := s.OpenLibrary.SearchBooks(
		ctx,
		request.Query,
		&request.Title,
		&request.Author,
		&request.Subject,
	)
	if err != nil {
		switch {
		case errors.Is(err, openlibrary.ErrBadRequest):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, openlibrary.ErrNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		case errors.Is(err, openlibrary.ErrInternalServer):
			return nil, status.Error(codes.Internal, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	books := make([]*betterreads.Book, 0)
	for _, book := range searchResult.Books {
		books = append(books, &betterreads.Book{
			Id:            book.CoverEditionKey,
			Title:         book.Title,
			AuthorName:    book.AuthorName,
			AuthorId:      book.AuthorKey,
			BookImage:     book.CoverImage,
			PublishedYear: int32(book.PublishYear), //nolint:gosec // integer overflow unlikely
			Isbn:          book.ISBN,
			RatingCount:   int32(book.RatingCount), //nolint:gosec // integer overflow unlikely
			RatingAverage: float32(book.RatingAverage),
			Source:        betterreads.BookSource_BOOK_SOURCE_OPEN_LIBRARY,
		})
	}
	return &betterreads.SearchBooksResponse{
		Books: books,
	}, nil
}

func verifySearchBooksRequest(request *betterreads.SearchBooksRequest) error {
	if request.Query == "" &&
		request.Title == "" &&
		request.Author == "" &&
		request.Subject == "" {
		return status.Error(codes.InvalidArgument, "must pass one search parameter")
	}
	return nil
}
