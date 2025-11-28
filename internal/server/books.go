package server

import (
	"context"
	"errors"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

// (GET /api/v1/books).
// SearchBooks implements betterreads.BetterReadsServiceServer
func (s *Server) SearchBooks(
	ctx context.Context,
	request *betterreads.SearchBooksRequest,
) (*betterreads.SearchBooksResponse, error) {
	if err := verifySearchBooksRequest(request); err != nil {
		return nil, err // TODO: Return gRPC status error
	}

	searchResult, err := s.OpenLibrary.SearchBooks(
		ctx,
		request.Query,
		&request.Title,
		&request.Author,
		&request.Subject,
	)
	if err != nil {
		// TODO: Map errors to gRPC codes
		return nil, err
	}

	books := make([]*betterreads.Book, 0)
	for _, book := range searchResult.Books {
		books = append(books, &betterreads.Book{
			Id:            book.CoverEditionKey,
			Title:         book.Title,
			AuthorName:    book.AuthorName,
			AuthorId:      book.AuthorKey,
			BookImage:     book.CoverImage,
			PublishedYear: int32(book.PublishYear),
			Isbn:          book.ISBN,
			RatingCount:   int32(book.RatingCount),
			RatingAverage: float32(book.RatingAverage),
			Source:        "OpenLibrary", // TODO: Use enum if defined
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
		return errors.New("must pass one search parameter")
	}
	return nil
}
