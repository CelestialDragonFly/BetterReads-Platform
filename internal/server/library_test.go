package server

import (
	"context"
	"errors"
	"testing"
	"time"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/data"
	"github.com/celestialdragonfly/betterreads/internal/headers"
	"github.com/celestialdragonfly/betterreads/internal/postgres"
	"github.com/celestialdragonfly/betterreads/internal/postgres/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_RemoveLibraryBook(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testBookID := "OL456M"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.RemoveLibraryBookRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful removal",
			ctx:  ctx,
			request: &betterreads.RemoveLibraryBookRequest{
				BookId: testBookID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					RemoveLibraryBook(gomock.Any(), testUserID, testBookID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.RemoveLibraryBookRequest{
				BookId: testBookID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty book_id",
			ctx:  ctx,
			request: &betterreads.RemoveLibraryBookRequest{
				BookId: "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "book not found",
			ctx:  ctx,
			request: &betterreads.RemoveLibraryBookRequest{
				BookId: testBookID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					RemoveLibraryBook(gomock.Any(), testUserID, testBookID).
					Return(postgres.ErrBookNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.RemoveLibraryBookRequest{
				BookId: testBookID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					RemoveLibraryBook(gomock.Any(), testUserID, testBookID).
					Return(errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}
			resp, err := s.RemoveLibraryBook(tt.ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok, "error should be a status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestServer_UpdateLibraryBook(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testBookID := "OL456M"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	validRequest := &betterreads.UpdateLibraryBookRequest{
		BookId:        testBookID,
		Title:         "Test Book",
		AuthorName:    "Test Author",
		BookImage:     "https://covers.openlibrary.org/b/olid/OL456M-L.jpg",
		Rating:        betterreads.BookRating(4),
		Source:        betterreads.BookSource_BOOK_SOURCE_OPEN_LIBRARY,
		ReadingStatus: betterreads.ReadingStatus_READING_STATUS_READ,
		ShelfIds:      []string{"shelf-1"},
	}

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.UpdateLibraryBookRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name:    "successful update",
			ctx:     ctx,
			request: validRequest,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateLibraryBook(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name:      "missing user_id in context",
			ctx:       ctxNoUserID,
			request:   validRequest,
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty book_id",
			ctx:  ctx,
			request: &betterreads.UpdateLibraryBookRequest{
				BookId:     "",
				Title:      "Test Book",
				AuthorName: "Test Author",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "empty title",
			ctx:  ctx,
			request: &betterreads.UpdateLibraryBookRequest{
				BookId:        testBookID,
				Title:         "",
				AuthorName:    "Test Author",
				ReadingStatus: betterreads.ReadingStatus_READING_STATUS_READ,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "empty author_name",
			ctx:  ctx,
			request: &betterreads.UpdateLibraryBookRequest{
				BookId:        testBookID,
				Title:         "Test Book",
				AuthorName:    "",
				ReadingStatus: betterreads.ReadingStatus_READING_STATUS_READ,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "rating below minimum",
			ctx:  ctx,
			request: &betterreads.UpdateLibraryBookRequest{
				BookId:        testBookID,
				Title:         "Test Book",
				AuthorName:    "Test Author",
				Rating:        betterreads.BookRating(-1),
				ReadingStatus: betterreads.ReadingStatus_READING_STATUS_READ,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "rating above maximum",
			ctx:  ctx,
			request: &betterreads.UpdateLibraryBookRequest{
				BookId:        testBookID,
				Title:         "Test Book",
				AuthorName:    "Test Author",
				Rating:        betterreads.BookRating(6),
				ReadingStatus: betterreads.ReadingStatus_READING_STATUS_READ,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "invalid source",
			ctx:  ctx,
			request: &betterreads.UpdateLibraryBookRequest{
				BookId:        testBookID,
				Title:         "Test Book",
				AuthorName:    "Test Author",
				Rating:        betterreads.BookRating(3),
				Source:        betterreads.BookSource(99),
				ReadingStatus: betterreads.ReadingStatus_READING_STATUS_READ,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name:    "database error",
			ctx:     ctx,
			request: validRequest,
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateLibraryBook(gomock.Any(), gomock.Any()).
					Return(errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}
			resp, err := s.UpdateLibraryBook(tt.ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok, "error should be a status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestServer_GetUserLibrary(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	testShelves := []*data.Shelf{
		{
			ID:        "shelf-1",
			Name:      "Currently Reading",
			UserID:    testUserID,
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
		{
			ID:        "shelf-2",
			Name:      "Want to Read",
			UserID:    testUserID,
			CreatedAt: testTime,
			UpdatedAt: testTime,
		},
	}

	shelvedBook := data.LibraryBook{
		BookID:     "OL456M",
		Title:      "Shelved Book",
		AuthorName: "Author One",
		BookImage:  "https://covers.openlibrary.org/b/olid/OL456M-L.jpg",
		Rating:     4,
		Source:     int32(betterreads.BookSource_BOOK_SOURCE_OPEN_LIBRARY),
		ShelfIDs:   []string{"shelf-1"},
		AddedAt:    testTime,
		UpdatedAt:  testTime,
	}

	unshelvedBook := data.LibraryBook{
		BookID:     "OL789M",
		Title:      "Unshelved Book",
		AuthorName: "Author Two",
		BookImage:  "https://covers.openlibrary.org/b/olid/OL789M-L.jpg",
		Rating:     3,
		Source:     int32(betterreads.BookSource_BOOK_SOURCE_OPEN_LIBRARY),
		ShelfIDs:   []string{},
		AddedAt:    testTime,
		UpdatedAt:  testTime,
	}

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.GetUserLibraryRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
		verify    func(*testing.T, *betterreads.GetUserLibraryResponse)
	}{
		{
			name: "successful retrieval with own user_id",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(testShelves, nil)
				m.EXPECT().
					GetUserLibrary(gomock.Any(), testUserID).
					Return([]*data.LibraryBook{&shelvedBook, &unshelvedBook}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserLibraryResponse) {
				t.Helper()
				assert.Len(t, resp.Shelves, 2)
				assert.Len(t, resp.UnshelvedBooks, 1)
				assert.Equal(t, "OL789M", resp.UnshelvedBooks[0].BookId)
				assert.Equal(t, int32(2), resp.Pagination.Total)
			},
		},
		{
			name: "successful retrieval with empty user_id falls back to authenticated user",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: "",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(testShelves, nil)
				m.EXPECT().
					GetUserLibrary(gomock.Any(), testUserID).
					Return([]*data.LibraryBook{}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserLibraryResponse) {
				t.Helper()
				assert.Len(t, resp.Shelves, 2)
				assert.Empty(t, resp.UnshelvedBooks)
				assert.Equal(t, int32(0), resp.Pagination.Total)
			},
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "permission denied when requesting another user's library",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: "other-user-456",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.PermissionDenied,
			wantErr:   true,
		},
		{
			name: "get shelves database error",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(nil, errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
		{
			name: "get library database error",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(testShelves, nil)
				m.EXPECT().
					GetUserLibrary(gomock.Any(), testUserID).
					Return(nil, errors.New("database connection failed"))
			},
			wantCode: codes.Internal,
			wantErr:  true,
		},
		{
			name: "shelved books are placed into correct shelves",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(testShelves, nil)
				m.EXPECT().
					GetUserLibrary(gomock.Any(), testUserID).
					Return([]*data.LibraryBook{&shelvedBook}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserLibraryResponse) {
				t.Helper()
				require.Len(t, resp.Shelves, 2)

				// shelf-1 should have the book
				shelf1 := resp.Shelves[0]
				assert.Equal(t, "shelf-1", shelf1.Shelf.Id)
				assert.Len(t, shelf1.Books, 1)
				assert.Equal(t, "OL456M", shelf1.Books[0].BookId)

				// shelf-2 should be empty
				shelf2 := resp.Shelves[1]
				assert.Equal(t, "shelf-2", shelf2.Shelf.Id)
				assert.Empty(t, shelf2.Books)

				assert.Empty(t, resp.UnshelvedBooks)
			},
		},
		{
			name: "empty library with no shelves",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return([]*data.Shelf{}, nil)
				m.EXPECT().
					GetUserLibrary(gomock.Any(), testUserID).
					Return([]*data.LibraryBook{}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserLibraryResponse) {
				t.Helper()
				assert.Empty(t, resp.Shelves)
				assert.Empty(t, resp.UnshelvedBooks)
				assert.Equal(t, int32(0), resp.Pagination.Total)
			},
		},
		{
			name: "pagination metadata is correctly populated",
			ctx:  ctx,
			request: &betterreads.GetUserLibraryRequest{
				UserId: testUserID,
				Page:   2,
				Limit:  10,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return([]*data.Shelf{}, nil)
				m.EXPECT().
					GetUserLibrary(gomock.Any(), testUserID).
					Return([]*data.LibraryBook{&unshelvedBook}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserLibraryResponse) {
				t.Helper()
				assert.Equal(t, int32(1), resp.Pagination.Total)
				assert.Equal(t, int32(2), resp.Pagination.Page)
				assert.Equal(t, int32(10), resp.Pagination.Limit)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}
			resp, err := s.GetUserLibrary(tt.ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok, "error should be a status error")
				assert.Equal(t, tt.wantCode, st.Code())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)

			if tt.verify != nil {
				tt.verify(t, resp)
			}
		})
	}
}
