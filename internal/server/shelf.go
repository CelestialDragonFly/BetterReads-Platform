package server

import (
	"context"
	"errors"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) CreateShelf(ctx context.Context, req *betterreads.CreateShelfRequest) (*betterreads.CreateShelfResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf name cannot be empty")
	}

	shelfID := uuid.New().String()
	now := time.Now()
	shelf := &data.Shelf{
		ID:        shelfID,
		Name:      req.Name,
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdShelf, err := s.DB.CreateShelf(ctx, shelf)
	if err != nil {
		if errors.Is(err, postgres.ErrShelfNameExists) {
			return nil, status.Error(codes.AlreadyExists, "shelf with this name already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create shelf: %v", err)
	}

	return &betterreads.CreateShelfResponse{
		Shelf: &betterreads.Shelf{
			Id:        createdShelf.ID,
			Name:      createdShelf.Name,
			UserId:    createdShelf.UserID,
			CreatedAt: timestamppb.New(createdShelf.CreatedAt),
			UpdatedAt: timestamppb.New(createdShelf.UpdatedAt),
		},
	}, nil
}

func (s *Server) UpdateShelf(ctx context.Context, req *betterreads.UpdateShelfRequest) (*betterreads.UpdateShelfResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.ShelfId == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf id is required")
	}

	// Ideally we check ownership here or in DB.
	// API UpdateShelf expects full object or partial?
	// DB UpdateShelf updates fields provided.

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf name cannot be empty")
	}

	shelf := &data.Shelf{
		ID:        req.ShelfId,
		UserID:    userID,
		Name:      req.Name,
		UpdatedAt: time.Now(),
	}

	updatedShelf, err := s.DB.UpdateShelf(ctx, shelf)
	if err != nil {
		if errors.Is(err, postgres.ErrShelfNotFound) {
			return nil, status.Error(codes.NotFound, "shelf not found")
		}
		if errors.Is(err, postgres.ErrShelfNameExists) {
			return nil, status.Error(codes.AlreadyExists, "shelf name already exists")
		}
		if errors.Is(err, postgres.ErrCannotUpdateDefaultShelf) {
			return nil, status.Error(codes.PermissionDenied, "cannot update default shelf")
		}
		return nil, status.Errorf(codes.Internal, "failed to update shelf: %v", err)
	}

	return &betterreads.UpdateShelfResponse{
		Shelf: &betterreads.Shelf{
			Id:        updatedShelf.ID,
			Name:      updatedShelf.Name,
			UserId:    updatedShelf.UserID,
			CreatedAt: timestamppb.New(updatedShelf.CreatedAt),
			UpdatedAt: timestamppb.New(updatedShelf.UpdatedAt),
		},
	}, nil
}

func (s *Server) DeleteShelf(ctx context.Context, req *betterreads.DeleteShelfRequest) (*betterreads.DeleteShelfResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.ShelfId == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf id is required")
	}

	if err := s.DB.DeleteShelf(ctx, userID, req.ShelfId); err != nil {
		if errors.Is(err, postgres.ErrShelfNotFound) {
			return nil, status.Error(codes.NotFound, "shelf not found")
		}
		if errors.Is(err, postgres.ErrCannotDeleteDefaultShelf) {
			return nil, status.Error(codes.PermissionDenied, "cannot delete default shelf")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete shelf: %v", err)
	}

	return &betterreads.DeleteShelfResponse{}, nil
}

func (s *Server) GetUserShelves(ctx context.Context, req *betterreads.GetUserShelvesRequest) (*betterreads.GetUserShelvesResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	targetUserID := req.UserId
	if targetUserID == "" {
		targetUserID = userID
	}

	// Only allow users to view their own shelves (privacy protection)
	if targetUserID != userID {
		return nil, status.Error(codes.PermissionDenied, "you can only view your own shelves")
	}

	shelves, err := s.DB.GetUserShelves(ctx, targetUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user shelves: %v", err)
	}

	var pbShelves []*betterreads.Shelf
	for _, shelf := range shelves {
		pbShelves = append(pbShelves, &betterreads.Shelf{
			Id:        shelf.ID,
			Name:      shelf.Name,
			UserId:    shelf.UserID,
			CreatedAt: timestamppb.New(shelf.CreatedAt),
			UpdatedAt: timestamppb.New(shelf.UpdatedAt),
		})
	}

	return &betterreads.GetUserShelvesResponse{
		Shelves: pbShelves,
	}, nil
}

func (s *Server) GetShelfBooks(ctx context.Context, req *betterreads.GetShelfBooksRequest) (*betterreads.GetShelfBooksResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.ShelfId == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf_id is required")
	}

	books, err := s.DB.GetShelfBooks(ctx, userID, req.ShelfId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get shelf books: %v", err)
	}

	// Sort logic not implemented in DB (except maybe default).
	// req.Sort is available. DB query relies on default.
	// Pagination (limit/page) not implemented in DB.

	var pbBooks []*betterreads.LibraryBook
	for _, b := range books {
		pbBooks = append(pbBooks, &betterreads.LibraryBook{
			AuthorName: b.AuthorName,
			BookId:     b.BookID,
			BookImage:  b.BookImage,
			Rating:     betterreads.BookRating(b.Rating),
			ShelfIds:   b.ShelfIDs,
			Source:     betterreads.BookSource(b.Source),
			Title:      b.Title,
			AddedAt:    timestamppb.New(b.AddedAt),
			UpdatedAt:  timestamppb.New(b.UpdatedAt),
		})
	}

	return &betterreads.GetShelfBooksResponse{
		Books: pbBooks,
		Pagination: &betterreads.PaginationMetadata{
			Total: int32(len(books)), //nolint:gosec // G115: len(books) is unlikely to overflow int32
			Page:  req.Page,
			Limit: req.Limit,
		},
	}, nil
}

func (s *Server) AddBookToShelf(ctx context.Context, req *betterreads.AddBookToShelfRequest) (*betterreads.AddBookToShelfResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book_id is required")
	}

	if req.ShelfId == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf_id is required")
	}

	if err := s.DB.AddBookToShelf(ctx, userID, req.BookId, req.ShelfId); err != nil {
		if errors.Is(err, postgres.ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found in library")
		}
		if errors.Is(err, postgres.ErrShelfNotFound) {
			return nil, status.Error(codes.NotFound, "shelf not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to add book to shelf: %v", err)
	}

	return &betterreads.AddBookToShelfResponse{}, nil
}

func (s *Server) RemoveBookFromShelf(ctx context.Context, req *betterreads.RemoveBookFromShelfRequest) (*betterreads.RemoveBookFromShelfResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book_id is required")
	}

	if req.ShelfId == "" {
		return nil, status.Error(codes.InvalidArgument, "shelf_id is required")
	}

	if err := s.DB.RemoveBookFromShelf(ctx, userID, req.BookId, req.ShelfId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to remove book from shelf: %v", err)
	}

	return &betterreads.RemoveBookFromShelfResponse{}, nil
}
