package server

import (
	"context"
	"errors"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RemoveLibraryBook.
func (s *Server) RemoveLibraryBook(ctx context.Context, req *betterreads.RemoveLibraryBookRequest) (*betterreads.RemoveLibraryBookResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book_id is required")
	}

	if err := s.DB.RemoveLibraryBook(ctx, userID, req.BookId); err != nil {
		if errors.Is(err, postgres.ErrBookNotFound) {
			return nil, status.Error(codes.NotFound, "book not found in library")
		}
		return nil, status.Errorf(codes.Internal, "failed to remove library book: %v", err)
	}

	return &betterreads.RemoveLibraryBookResponse{}, nil
}

// UpdateLibraryBook.
func (s *Server) UpdateLibraryBook(ctx context.Context, req *betterreads.UpdateLibraryBookRequest) (*betterreads.UpdateLibraryBookResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	if req.BookId == "" {
		return nil, status.Error(codes.InvalidArgument, "book_id is required")
	}

	if req.Title == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	if req.AuthorName == "" {
		return nil, status.Error(codes.InvalidArgument, "author_name is required")
	}

	// Validate rating range (0 = unspecified, 1-5 = actual ratings)
	if req.Rating < 0 || req.Rating > 5 {
		return nil, status.Error(codes.InvalidArgument, "rating must be between 0 and 5")
	}

	// Validate source enum
	if req.Source < 0 || req.Source > 3 {
		return nil, status.Error(codes.InvalidArgument, "invalid book source")
	}

	// Validate reading status enum
	if req.ReadingStatus == betterreads.ReadingStatus_READING_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "reading status must be specified")
	}
	if req.ReadingStatus < 0 || req.ReadingStatus > 4 {
		return nil, status.Error(codes.InvalidArgument, "invalid reading status")
	}

	now := time.Now()
	book := &data.LibraryBook{
		UserID:        userID,
		BookID:        req.BookId,
		Title:         req.Title,
		AuthorName:    req.AuthorName,
		BookImage:     req.BookImage,
		Rating:        int32(req.Rating),
		Source:        int32(req.Source),
		ReadingStatus: int32(req.ReadingStatus),
		ShelfIDs:      req.ShelfIds,
		AddedAt:       now, // Used only on INSERT, preserved on UPDATE (see DB upsert query)
		UpdatedAt:     now, // Always updated on both INSERT and UPDATE
	}

	if err := s.DB.UpdateLibraryBook(ctx, book); err != nil {
		// UpdateLibraryBook could fail if shelf IDs are invalid, but we don't have a specific error for that yet in postgres package usually.
		return nil, status.Errorf(codes.Internal, "failed to update library book: %v", err)
	}

	return &betterreads.UpdateLibraryBookResponse{}, nil
}

// GetUserLibrary.
func (s *Server) GetUserLibrary(ctx context.Context, req *betterreads.GetUserLibraryRequest) (*betterreads.GetUserLibraryResponse, error) {
	userID, ok := headers.GetUserID(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user not found in context")
	}

	targetUserID := req.UserId
	if targetUserID == "" {
		targetUserID = userID
	}

	// Only allow users to view their own library (privacy protection)
	if targetUserID != userID {
		return nil, status.Error(codes.PermissionDenied, "you can only view your own library")
	}

	// 1. Get all shelves
	shelves, err := s.DB.GetUserShelves(ctx, targetUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user shelves: %v", err)
	}

	// 2. Get all books
	books, err := s.DB.GetUserLibrary(ctx, targetUserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user library: %v", err)
	}

	// 3. Construct response

	// Map shelfID -> Schema
	shelfMap := make(map[string]*betterreads.ShelfWithBooks)

	// Initialize shelves in map
	for _, s := range shelves {
		shelfMap[s.ID] = &betterreads.ShelfWithBooks{
			Shelf: &betterreads.Shelf{
				Id:        s.ID,
				Name:      s.Name,
				UserId:    s.UserID,
				CreatedAt: timestamppb.New(s.CreatedAt),
				UpdatedAt: timestamppb.New(s.UpdatedAt),
			},
			Books: []*betterreads.LibraryBook{},
		}
	}

	var unshelved []*betterreads.LibraryBook

	for _, b := range books {
		pbBook := &betterreads.LibraryBook{
			AuthorName:    b.AuthorName,
			BookId:        b.BookID,
			BookImage:     b.BookImage,
			Rating:        betterreads.BookRating(b.Rating),
			ShelfIds:      b.ShelfIDs,
			Source:        betterreads.BookSource(b.Source),
			ReadingStatus: betterreads.ReadingStatus(b.ReadingStatus),
			Title:         b.Title,
			AddedAt:       timestamppb.New(b.AddedAt),
			UpdatedAt:     timestamppb.New(b.UpdatedAt),
		}

		if len(b.ShelfIDs) == 0 {
			unshelved = append(unshelved, pbBook)
		} else {
			for _, sid := range b.ShelfIDs {
				if shelf, exists := shelfMap[sid]; exists {
					shelf.Books = append(shelf.Books, pbBook)
				}
				// If shelf doesn't exist (maybe deleted but book still references it?), ignore or create dummy?
				// Should not happen with FK constraints.
			}
		}
	}

	// Convert map to slice
	var finalShelves []*betterreads.ShelfWithBooks
	// Maintain order? Shelves were from GetUserShelves (ordered by CreatedAt).
	// So iterate original shelves list.
	for _, s := range shelves {
		if swb, ok := shelfMap[s.ID]; ok {
			finalShelves = append(finalShelves, swb)
		}
	}

	return &betterreads.GetUserLibraryResponse{
		Shelves:        finalShelves,
		UnshelvedBooks: unshelved,
		Pagination: &betterreads.PaginationMetadata{
			Total: int32(len(books)), //nolint:gosec // G115: len(books) is unlikely to overflow int32
			Page:  req.Page,
			Limit: req.Limit,
		},
	}, nil
}
