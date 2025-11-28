package server

import (
	"context"
	"errors"
	"testing"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockOpenLibraryClient is a mock implementation of the OpenLibraryClient interface.
type MockOpenLibraryClient struct {
	mock.Mock
}

// SearchBooks is a mock implementation of the SearchBooks method.
func (m *MockOpenLibraryClient) SearchBooks(ctx context.Context, query string, _, _, _ *string) (*openlibrary.SearchBooksResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*openlibrary.SearchBooksResponse), args.Error(1)
}

// NewMockClient creates a new mock OpenLibrary client.
func NewMockClient() *MockOpenLibraryClient {
	return &MockOpenLibraryClient{}
}

func TestServer_SearchBooks_Success(t *testing.T) {
	t.Parallel()
	// Setup
	mockClient := NewMockClient()
	server := NewServer(&Config{
		OpenLibrary: mockClient,
	})

	// Test data
	testQuery := "test query"
	mockResponse := &openlibrary.SearchBooksResponse{
		Books: []openlibrary.Book{
			{
				AuthorKey:       "OL123A",
				AuthorName:      "Test Author",
				CoverEditionKey: "OL456M",
				CoverImage:      "https://covers.openlibrary.org/b/olid/OL456M-L.jpg",
				ISBN:            "1234567890",
				Title:           "Test Book",
				RatingAverage:   4.5,
				RatingCount:     100,
				PublishYear:     2020,
				Source:          string(betterreads.BookSource_OPEN_LIBRARY),
			},
		},
	}

	// Expectations
	mockClient.On("SearchBooks", mock.Anything, testQuery).Return(mockResponse, nil)

	// Execute
	resp, err := server.SearchBooks(context.Background(), &betterreads.SearchBooksRequest{
		Query: testQuery,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	// Verify response content
	assert.Len(t, resp.Books, 1)
	assert.Equal(t, "OL456M", resp.Books[0].Id)
	assert.Equal(t, "Test Author", resp.Books[0].AuthorName)
	assert.Equal(t, "OL123A", resp.Books[0].AuthorId)
	assert.Equal(t, "Test Book", resp.Books[0].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL456M-L.jpg", resp.Books[0].BookImage)
	assert.Equal(t, "1234567890", resp.Books[0].Isbn)
	assert.Equal(t, float32(4.5), resp.Books[0].RatingAverage)
	assert.Equal(t, int32(100), resp.Books[0].RatingCount)
	assert.Equal(t, int32(2020), resp.Books[0].PublishedYear)
	assert.Equal(t, betterreads.BookSource_OPEN_LIBRARY, resp.Books[0].Source)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestServer_SearchBooks_NilQuery(t *testing.T) {
	t.Parallel()
	// Setup
	mockClient := NewMockClient()
	server := NewServer(&Config{
		OpenLibrary: mockClient,
	})

	// Execute
	resp, err := server.SearchBooks(context.Background(), &betterreads.SearchBooksRequest{
		Query: "", // Empty query should trigger an error
	})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())

	// Verify no calls were made to the mock
	mockClient.AssertNotCalled(t, "SearchBooks")
}

func TestServer_SearchBooks_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		error        error
		expectedCode codes.Code
	}{
		{
			name:         "Bad Request Error",
			error:        openlibrary.ErrBadRequest,
			expectedCode: codes.InvalidArgument,
		},
		{
			name:         "Not Found Error",
			error:        openlibrary.ErrNotFound,
			expectedCode: codes.NotFound,
		},
		{
			name:         "Internal Server Error",
			error:        openlibrary.ErrInternalServer,
			expectedCode: codes.Internal,
		},
		{
			name:         "Unknown Error",
			error:        errors.New("unknown error"),
			expectedCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Setup
			mockClient := NewMockClient()
			server := NewServer(&Config{
				OpenLibrary: mockClient,
			})

			// Test data
			testQuery := "test query"

			// Expectations
			mockClient.On("SearchBooks", mock.Anything, testQuery).Return(nil, tt.error)

			// Execute
			resp, err := server.SearchBooks(context.Background(), &betterreads.SearchBooksRequest{
				Query: testQuery,
			})

			// Assert
			assert.Error(t, err)
			assert.Nil(t, resp)
			st, ok := status.FromError(err)
			assert.True(t, ok)
			assert.Equal(t, tt.expectedCode, st.Code())

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestServer_SearchBooks_EmptyResults(t *testing.T) {
	t.Parallel()
	// Setup
	mockClient := NewMockClient()
	server := NewServer(&Config{
		OpenLibrary: mockClient,
	})

	// Test data
	testQuery := "test query"
	mockResponse := &openlibrary.SearchBooksResponse{
		Books: []openlibrary.Book{}, // Empty results
	}

	// Expectations
	mockClient.On("SearchBooks", mock.Anything, testQuery).Return(mockResponse, nil)

	// Execute
	resp, err := server.SearchBooks(context.Background(), &betterreads.SearchBooksRequest{
		Query: testQuery,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Empty(t, resp.Books)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestServer_SearchBooks_MultipleBooks(t *testing.T) {
	t.Parallel()
	// Setup
	mockClient := NewMockClient()
	server := NewServer(&Config{
		OpenLibrary: mockClient,
	})

	// Test data
	testQuery := "multiple books"
	mockResponse := &openlibrary.SearchBooksResponse{
		Books: []openlibrary.Book{
			{
				AuthorKey:       "OL123A",
				AuthorName:      "Author One",
				CoverEditionKey: "OL456M",
				CoverImage:      "https://covers.openlibrary.org/b/olid/OL456M-L.jpg",
				ISBN:            "1234567890",
				Title:           "First Book",
				RatingAverage:   4.5,
				RatingCount:     100,
				PublishYear:     2020,
				Source:          string(betterreads.BookSource_OPEN_LIBRARY),
			},
			{
				AuthorKey:       "OL789A",
				AuthorName:      "Author Two",
				CoverEditionKey: "OL101M",
				CoverImage:      "https://covers.openlibrary.org/b/olid/OL101M-L.jpg",
				ISBN:            "0987654321",
				Title:           "Second Book",
				RatingAverage:   3.8,
				RatingCount:     75,
				PublishYear:     2018,
				Source:          string(betterreads.BookSource_OPEN_LIBRARY),
			},
			{
				AuthorKey:       "OL456A",
				AuthorName:      "Author Three",
				CoverEditionKey: "OL202M",
				CoverImage:      "https://covers.openlibrary.org/b/olid/OL202M-L.jpg",
				ISBN:            "5678901234",
				Title:           "Third Book",
				RatingAverage:   4.2,
				RatingCount:     50,
				PublishYear:     2022,
				Source:          string(betterreads.BookSource_OPEN_LIBRARY),
			},
		},
	}

	// Expectations
	mockClient.On("SearchBooks", mock.Anything, testQuery).Return(mockResponse, nil)

	// Execute
	resp, err := server.SearchBooks(context.Background(), &betterreads.SearchBooksRequest{
		Query: testQuery,
	})

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Books, 3)

	// Verify first book
	assert.Equal(t, "OL456M", resp.Books[0].Id)
	assert.Equal(t, "Author One", resp.Books[0].AuthorName)
	assert.Equal(t, "OL123A", resp.Books[0].AuthorId)
	assert.Equal(t, "First Book", resp.Books[0].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL456M-L.jpg", resp.Books[0].BookImage)
	assert.Equal(t, "1234567890", resp.Books[0].Isbn)
	assert.Equal(t, float32(4.5), resp.Books[0].RatingAverage)
	assert.Equal(t, int32(100), resp.Books[0].RatingCount)
	assert.Equal(t, int32(2020), resp.Books[0].PublishedYear)
	assert.Equal(t, betterreads.BookSource_OPEN_LIBRARY, resp.Books[0].Source)

	// Verify second book
	assert.Equal(t, "OL101M", resp.Books[1].Id)
	assert.Equal(t, "Author Two", resp.Books[1].AuthorName)
	assert.Equal(t, "OL789A", resp.Books[1].AuthorId)
	assert.Equal(t, "Second Book", resp.Books[1].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL101M-L.jpg", resp.Books[1].BookImage)
	assert.Equal(t, "0987654321", resp.Books[1].Isbn)
	assert.Equal(t, float32(3.8), resp.Books[1].RatingAverage)
	assert.Equal(t, int32(75), resp.Books[1].RatingCount)
	assert.Equal(t, int32(2018), resp.Books[1].PublishedYear)
	assert.Equal(t, betterreads.BookSource_OPEN_LIBRARY, resp.Books[1].Source)

	// Verify third book
	assert.Equal(t, "OL202M", resp.Books[2].Id)
	assert.Equal(t, "Author Three", resp.Books[2].AuthorName)
	assert.Equal(t, "OL456A", resp.Books[2].AuthorId)
	assert.Equal(t, "Third Book", resp.Books[2].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL202M-L.jpg", resp.Books[2].BookImage)
	assert.Equal(t, "5678901234", resp.Books[2].Isbn)
	assert.Equal(t, float32(4.2), resp.Books[2].RatingAverage)
	assert.Equal(t, int32(50), resp.Books[2].RatingCount)
	assert.Equal(t, int32(2022), resp.Books[2].PublishedYear)
	assert.Equal(t, betterreads.BookSource_OPEN_LIBRARY, resp.Books[2].Source)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}
