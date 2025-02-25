package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

type Config struct {
}

type Server struct {
}

var _ betterreads.StrictServerInterface = (*Server)(nil)

func NewServer(_ *Config) *Server {
	return &Server{}
}

// GetBooks implements betterreads.StrictServerInterface.
func (s *Server) GetApiV1Books(ctx context.Context, request betterreads.GetApiV1BooksRequestObject) (betterreads.GetApiV1BooksResponseObject, error) {
	return betterreads.GetApiV1Books200JSONResponse{
		Books: []betterreads.Book{
			{
				Author:      "George Orwell",
				BookImage:   "https://example.com/1984_cover.jpg",
				Description: "1984 is a dystopian novel set in a totalitarian society under constant surveillance, where the protagonist, Winston Smith, struggles to assert his individuality.",
				Genre:       "Dystopian, Political Fiction",
				Id:          "1",
				Title:       "1984",
			},
		},
	}, nil
}
