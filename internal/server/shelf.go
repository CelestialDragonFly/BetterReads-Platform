package server

import (
	"context"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
)

func (s *Server) CreateShelf(_ context.Context, _ *betterreads.CreateShelfRequest) (*betterreads.CreateShelfResponse, error) {
	return &betterreads.CreateShelfResponse{}, nil
}

func (s *Server) UpdateShelf(_ context.Context, _ *betterreads.UpdateShelfRequest) (*betterreads.UpdateShelfResponse, error) {
	return &betterreads.UpdateShelfResponse{}, nil
}

func (s *Server) DeleteShelf(_ context.Context, _ *betterreads.DeleteShelfRequest) (*betterreads.DeleteShelfResponse, error) {
	return &betterreads.DeleteShelfResponse{}, nil
}

func (s *Server) GetUserShelves(_ context.Context, _ *betterreads.GetUserShelvesRequest) (*betterreads.GetUserShelvesResponse, error) {
	return &betterreads.GetUserShelvesResponse{}, nil
}

func (s *Server) GetShelfBooks(_ context.Context, _ *betterreads.GetShelfBooksRequest) (*betterreads.GetShelfBooksResponse, error) {
	return &betterreads.GetShelfBooksResponse{}, nil
}

func (s *Server) AddBookToShelf(_ context.Context, _ *betterreads.AddBookToShelfRequest) (*betterreads.AddBookToShelfResponse, error) {
	return &betterreads.AddBookToShelfResponse{}, nil
}

func (s *Server) RemoveBookFromShelf(_ context.Context, _ *betterreads.RemoveBookFromShelfRequest) (*betterreads.RemoveBookFromShelfResponse, error) {
	return &betterreads.RemoveBookFromShelfResponse{}, nil
}
