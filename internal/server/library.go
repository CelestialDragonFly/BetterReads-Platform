package server

import (
	"context"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
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

	book := &data.LibraryBook{
		UserID:     userID,
		BookID:     req.BookId,
		Title:      req.Title,
		AuthorName: req.AuthorName,
		BookImage:  req.BookImage,
		Rating:     int32(req.Rating),
		Source:     int32(req.Source),
		ShelfIDs:   req.ShelfIds,
		AddedAt:    time.Now(), // Default, DB ignores if update? No, DB sets. My DB implementation sets both added_at and updated_at on insert, updates updated_at on conflict.
		// Wait, DB UpdateLibraryBook implementation:
		/*
		   INSERT INTO library_books (...) VALUES ($8, $9) ...
		   SET ... updated_at = EXCLUDED.updated_at
		*/
		// So I should set AddedAt and UpdatedAt here.
		UpdatedAt: time.Now(),
	}

	if err := s.DB.UpdateLibraryBook(ctx, book); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update library book: %v", err)
	}

	return &betterreads.UpdateLibraryBookResponse{}, nil
}

// GetUserLibrary.
func (s *Server) GetUserLibrary(ctx context.Context, req *betterreads.GetUserLibraryRequest) (*betterreads.GetUserLibraryResponse, error) {
	// req.UserId exists. If empty, assume current?
	targetUserID := req.UserId
	if targetUserID == "" {
		uid, ok := headers.GetUserID(ctx)
		if ok {
			targetUserID = uid
		} else {
			return nil, status.Error(codes.InvalidArgument, "user_id is required")
		}
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
			AuthorName: b.AuthorName,
			BookId:     b.BookID,
			BookImage:  b.BookImage,
			Rating:     betterreads.BookRating(b.Rating),
			ShelfIds:   b.ShelfIDs,
			Source:     betterreads.BookSource(b.Source),
			Title:      b.Title,
			AddedAt:    timestamppb.New(b.AddedAt),
			UpdatedAt:  timestamppb.New(b.UpdatedAt),
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
