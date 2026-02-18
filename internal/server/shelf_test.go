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

// testShelf returns a reusable data.Shelf for use across test cases.
func testShelf(id, name, userID string, t time.Time) *data.Shelf {
	return &data.Shelf{
		ID:        id,
		Name:      name,
		UserID:    userID,
		CreatedAt: t,
		UpdatedAt: t,
	}
}

// ---- CreateShelf -------------------------------------------------------

func TestServer_CreateShelf(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	returnedShelf := testShelf("shelf-abc", "Fantasy", testUserID, testTime)

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.CreateShelfRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
		verify    func(*testing.T, *betterreads.CreateShelfResponse)
	}{
		{
			name: "successful creation",
			ctx:  ctx,
			request: &betterreads.CreateShelfRequest{
				Name: "Fantasy",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					CreateShelf(gomock.Any(), gomock.Any()).
					Return(returnedShelf, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.CreateShelfResponse) {
				t.Helper()
				require.NotNil(t, resp.Shelf)
				assert.Equal(t, "shelf-abc", resp.Shelf.Id)
				assert.Equal(t, "Fantasy", resp.Shelf.Name)
				assert.Equal(t, testUserID, resp.Shelf.UserId)
			},
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.CreateShelfRequest{
				Name: "Fantasy",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty shelf name",
			ctx:  ctx,
			request: &betterreads.CreateShelfRequest{
				Name: "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "shelf name already exists",
			ctx:  ctx,
			request: &betterreads.CreateShelfRequest{
				Name: "Fantasy",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					CreateShelf(gomock.Any(), gomock.Any()).
					Return(nil, postgres.ErrShelfNameExists)
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.CreateShelfRequest{
				Name: "Fantasy",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					CreateShelf(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database connection failed"))
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
			resp, err := s.CreateShelf(tt.ctx, tt.request)

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

// ---- UpdateShelf -------------------------------------------------------

func TestServer_UpdateShelf(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testShelfID := "shelf-abc"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	returnedShelf := testShelf(testShelfID, "Science Fiction", testUserID, testTime)

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.UpdateShelfRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
		verify    func(*testing.T, *betterreads.UpdateShelfResponse)
	}{
		{
			name: "successful update",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "Science Fiction",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateShelf(gomock.Any(), gomock.Any()).
					Return(returnedShelf, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.UpdateShelfResponse) {
				t.Helper()
				require.NotNil(t, resp.Shelf)
				assert.Equal(t, testShelfID, resp.Shelf.Id)
				assert.Equal(t, "Science Fiction", resp.Shelf.Name)
				assert.Equal(t, testUserID, resp.Shelf.UserId)
			},
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "Science Fiction",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty shelf_id",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: "",
				Name:    "Science Fiction",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "empty shelf name",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "shelf not found",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "Science Fiction",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateShelf(gomock.Any(), gomock.Any()).
					Return(nil, postgres.ErrShelfNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "shelf name already exists",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "Science Fiction",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateShelf(gomock.Any(), gomock.Any()).
					Return(nil, postgres.ErrShelfNameExists)
			},
			wantCode: codes.AlreadyExists,
			wantErr:  true,
		},
		{
			name: "cannot update default shelf",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "Science Fiction",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateShelf(gomock.Any(), gomock.Any()).
					Return(nil, postgres.ErrCannotUpdateDefaultShelf)
			},
			wantCode: codes.PermissionDenied,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.UpdateShelfRequest{
				ShelfId: testShelfID,
				Name:    "Science Fiction",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					UpdateShelf(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("database connection failed"))
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
			resp, err := s.UpdateShelf(tt.ctx, tt.request)

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

// ---- DeleteShelf -------------------------------------------------------

func TestServer_DeleteShelf(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testShelfID := "shelf-abc"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.DeleteShelfRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful deletion",
			ctx:  ctx,
			request: &betterreads.DeleteShelfRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					DeleteShelf(gomock.Any(), testUserID, testShelfID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.DeleteShelfRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty shelf_id",
			ctx:  ctx,
			request: &betterreads.DeleteShelfRequest{
				ShelfId: "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "shelf not found",
			ctx:  ctx,
			request: &betterreads.DeleteShelfRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					DeleteShelf(gomock.Any(), testUserID, testShelfID).
					Return(postgres.ErrShelfNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "cannot delete default shelf",
			ctx:  ctx,
			request: &betterreads.DeleteShelfRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					DeleteShelf(gomock.Any(), testUserID, testShelfID).
					Return(postgres.ErrCannotDeleteDefaultShelf)
			},
			wantCode: codes.PermissionDenied,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.DeleteShelfRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					DeleteShelf(gomock.Any(), testUserID, testShelfID).
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
			resp, err := s.DeleteShelf(tt.ctx, tt.request)

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

// ---- GetUserShelves ----------------------------------------------------

func TestServer_GetUserShelves(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	dbShelves := []*data.Shelf{
		testShelf("shelf-1", "Currently Reading", testUserID, testTime),
		testShelf("shelf-2", "Want to Read", testUserID, testTime),
	}

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.GetUserShelvesRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
		verify    func(*testing.T, *betterreads.GetUserShelvesResponse)
	}{
		{
			name: "successful retrieval with explicit user_id",
			ctx:  ctx,
			request: &betterreads.GetUserShelvesRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(dbShelves, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserShelvesResponse) {
				t.Helper()
				require.Len(t, resp.Shelves, 2)
				assert.Equal(t, "shelf-1", resp.Shelves[0].Id)
				assert.Equal(t, "Currently Reading", resp.Shelves[0].Name)
				assert.Equal(t, testUserID, resp.Shelves[0].UserId)
				assert.Equal(t, "shelf-2", resp.Shelves[1].Id)
				assert.Equal(t, "Want to Read", resp.Shelves[1].Name)
			},
		},
		{
			name: "empty user_id falls back to authenticated user",
			ctx:  ctx,
			request: &betterreads.GetUserShelvesRequest{
				UserId: "",
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return(dbShelves, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserShelvesResponse) {
				t.Helper()
				assert.Len(t, resp.Shelves, 2)
			},
		},
		{
			name: "empty shelf list returned",
			ctx:  ctx,
			request: &betterreads.GetUserShelvesRequest{
				UserId: testUserID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetUserShelves(gomock.Any(), testUserID).
					Return([]*data.Shelf{}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetUserShelvesResponse) {
				t.Helper()
				assert.Empty(t, resp.Shelves)
			},
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.GetUserShelvesRequest{
				UserId: testUserID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "permission denied when requesting another user's shelves",
			ctx:  ctx,
			request: &betterreads.GetUserShelvesRequest{
				UserId: "other-user-456",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.PermissionDenied,
			wantErr:   true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.GetUserShelvesRequest{
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockAPI(ctrl)
			tt.setupMock(mockDB)

			s := &Server{DB: mockDB}
			resp, err := s.GetUserShelves(tt.ctx, tt.request)

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

// ---- GetShelfBooks -----------------------------------------------------

func TestServer_GetShelfBooks(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testShelfID := "shelf-abc"
	testTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	dbBooks := []*data.LibraryBook{
		{
			BookID:     "OL111M",
			Title:      "Dune",
			AuthorName: "Frank Herbert",
			BookImage:  "https://covers.openlibrary.org/b/olid/OL111M-L.jpg",
			Rating:     4,
			Source:     int32(betterreads.BookSource_BOOK_SOURCE_OPEN_LIBRARY),
			ShelfIDs:   []string{testShelfID},
			AddedAt:    testTime,
			UpdatedAt:  testTime,
		},
		{
			BookID:     "OL222M",
			Title:      "Foundation",
			AuthorName: "Isaac Asimov",
			BookImage:  "https://covers.openlibrary.org/b/olid/OL222M-L.jpg",
			Rating:     4,
			Source:     int32(betterreads.BookSource_BOOK_SOURCE_OPEN_LIBRARY),
			ShelfIDs:   []string{testShelfID},
			AddedAt:    testTime,
			UpdatedAt:  testTime,
		},
	}

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.GetShelfBooksRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
		verify    func(*testing.T, *betterreads.GetShelfBooksResponse)
	}{
		{
			name: "successful retrieval with multiple books",
			ctx:  ctx,
			request: &betterreads.GetShelfBooksRequest{
				ShelfId: testShelfID,
				Page:    1,
				Limit:   10,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetShelfBooks(gomock.Any(), testUserID, testShelfID).
					Return(dbBooks, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetShelfBooksResponse) {
				t.Helper()
				require.Len(t, resp.Books, 2)
				assert.Equal(t, "OL111M", resp.Books[0].BookId)
				assert.Equal(t, "Dune", resp.Books[0].Title)
				assert.Equal(t, "Frank Herbert", resp.Books[0].AuthorName)
				assert.Equal(t, "OL222M", resp.Books[1].BookId)
				assert.Equal(t, "Foundation", resp.Books[1].Title)
				assert.Equal(t, "Isaac Asimov", resp.Books[1].AuthorName)
			},
		},
		{
			name: "pagination metadata is correctly propagated",
			ctx:  ctx,
			request: &betterreads.GetShelfBooksRequest{
				ShelfId: testShelfID,
				Page:    3,
				Limit:   5,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetShelfBooks(gomock.Any(), testUserID, testShelfID).
					Return(dbBooks, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetShelfBooksResponse) {
				t.Helper()
				assert.Equal(t, int32(2), resp.Pagination.Total)
				assert.Equal(t, int32(3), resp.Pagination.Page)
				assert.Equal(t, int32(5), resp.Pagination.Limit)
			},
		},
		{
			name: "empty shelf returns no books",
			ctx:  ctx,
			request: &betterreads.GetShelfBooksRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetShelfBooks(gomock.Any(), testUserID, testShelfID).
					Return([]*data.LibraryBook{}, nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
			verify: func(t *testing.T, resp *betterreads.GetShelfBooksResponse) {
				t.Helper()
				assert.Empty(t, resp.Books)
				assert.Equal(t, int32(0), resp.Pagination.Total)
			},
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.GetShelfBooksRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty shelf_id",
			ctx:  ctx,
			request: &betterreads.GetShelfBooksRequest{
				ShelfId: "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.GetShelfBooksRequest{
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					GetShelfBooks(gomock.Any(), testUserID, testShelfID).
					Return(nil, errors.New("database connection failed"))
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
			resp, err := s.GetShelfBooks(tt.ctx, tt.request)

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

// ---- AddBookToShelf ----------------------------------------------------

func TestServer_AddBookToShelf(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testBookID := "OL456M"
	testShelfID := "shelf-abc"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.AddBookToShelfRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful addition",
			ctx:  ctx,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					AddBookToShelf(gomock.Any(), testUserID, testBookID, testShelfID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty book_id",
			ctx:  ctx,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  "",
				ShelfId: testShelfID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "empty shelf_id",
			ctx:  ctx,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  testBookID,
				ShelfId: "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "book not found in library",
			ctx:  ctx,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					AddBookToShelf(gomock.Any(), testUserID, testBookID, testShelfID).
					Return(postgres.ErrBookNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "shelf not found",
			ctx:  ctx,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					AddBookToShelf(gomock.Any(), testUserID, testBookID, testShelfID).
					Return(postgres.ErrShelfNotFound)
			},
			wantCode: codes.NotFound,
			wantErr:  true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.AddBookToShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					AddBookToShelf(gomock.Any(), testUserID, testBookID, testShelfID).
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
			resp, err := s.AddBookToShelf(tt.ctx, tt.request)

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

// ---- RemoveBookFromShelf -----------------------------------------------

func TestServer_RemoveBookFromShelf(t *testing.T) {
	t.Parallel()

	testUserID := "test-user-123"
	testBookID := "OL456M"
	testShelfID := "shelf-abc"
	ctx := context.WithValue(context.Background(), headers.UserIDContextKey, testUserID)
	ctxNoUserID := context.Background()

	tests := []struct {
		name      string
		ctx       context.Context
		request   *betterreads.RemoveBookFromShelfRequest
		setupMock func(*mocks.MockAPI)
		wantCode  codes.Code
		wantErr   bool
	}{
		{
			name: "successful removal",
			ctx:  ctx,
			request: &betterreads.RemoveBookFromShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					RemoveBookFromShelf(gomock.Any(), testUserID, testBookID, testShelfID).
					Return(nil)
			},
			wantCode: codes.OK,
			wantErr:  false,
		},
		{
			name: "missing user_id in context",
			ctx:  ctxNoUserID,
			request: &betterreads.RemoveBookFromShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.Unauthenticated,
			wantErr:   true,
		},
		{
			name: "empty book_id",
			ctx:  ctx,
			request: &betterreads.RemoveBookFromShelfRequest{
				BookId:  "",
				ShelfId: testShelfID,
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "empty shelf_id",
			ctx:  ctx,
			request: &betterreads.RemoveBookFromShelfRequest{
				BookId:  testBookID,
				ShelfId: "",
			},
			setupMock: func(_ *mocks.MockAPI) {},
			wantCode:  codes.InvalidArgument,
			wantErr:   true,
		},
		{
			name: "database error",
			ctx:  ctx,
			request: &betterreads.RemoveBookFromShelfRequest{
				BookId:  testBookID,
				ShelfId: testShelfID,
			},
			setupMock: func(m *mocks.MockAPI) {
				m.EXPECT().
					RemoveBookFromShelf(gomock.Any(), testUserID, testBookID, testShelfID).
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
			resp, err := s.RemoveBookFromShelf(tt.ctx, tt.request)

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
