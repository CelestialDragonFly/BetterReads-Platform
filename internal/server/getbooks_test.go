package server

import (
	"context"
	"errors"
	"testing"

	betterreads "github.com/celestialdragonfly/betterreads/generated"
	"github.com/celestialdragonfly/betterreads/internal/openlibrary"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestServer_GetApiV1Books_Success(t *testing.T) {
	t.Parallel()
	// Setup
	mockClient := NewMockClient()
	server := NewServer(&Config{
		OpenLibrary: mockClient, // Will be implicitly converted to OpenLibraryClient interface
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
				Source:          "open_library",
			},
		},
	}

	// Expectations
	mockClient.On("SearchBooks", mock.Anything, testQuery).Return(mockResponse, nil)

	// Execute
	queryPtr := testQuery
	resp, err := server.GetApiV1Books(context.Background(), betterreads.GetApiV1BooksRequestObject{
		Params: betterreads.GetApiV1BooksParams{
			Query: &queryPtr,
		},
	})

	// Assert
	assert.NoError(t, err)

	// Type assertion to get the specific response type
	successResp, ok := resp.(betterreads.GetApiV1Books200JSONResponse)
	assert.True(t, ok, "Expected a 200 response")

	// Verify response content
	assert.Len(t, successResp.Books, 1)
	assert.Equal(t, "OL456M", successResp.Books[0].Id)
	assert.Equal(t, "Test Author", successResp.Books[0].AuthorName)
	assert.Equal(t, "OL123A", successResp.Books[0].AuthorId)
	assert.Equal(t, "Test Book", successResp.Books[0].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL456M-L.jpg", successResp.Books[0].BookImage)
	assert.Equal(t, "1234567890", successResp.Books[0].Isbn)
	assert.Equal(t, float32(4.5), successResp.Books[0].RatingAverage)
	assert.Equal(t, 100, successResp.Books[0].RatingCount)
	assert.Equal(t, 2020, successResp.Books[0].PublishedYear)
	assert.Equal(t, "open_library", successResp.Books[0].Source)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestServer_GetApiV1Books_NilQuery(t *testing.T) {
	t.Parallel()
	// Setup
	mockClient := NewMockClient()
	server := NewServer(&Config{
		OpenLibrary: mockClient,
	})

	// Execute
	resp, err := server.GetApiV1Books(context.Background(), betterreads.GetApiV1BooksRequestObject{
		Params: betterreads.GetApiV1BooksParams{
			Query: nil, // nil query should trigger an error
		},
	})

	// Assert
	assert.NoError(t, err)

	// Type assertion to get the specific response type
	badRequestResp, ok := resp.(betterreads.GetApiV1Books400JSONResponse)
	assert.True(t, ok, "Expected a 400 response")

	// Verify response content
	assert.Equal(t, "BAD_REQUEST", badRequestResp.Code)
	assert.Equal(t, "search books - invalid request", badRequestResp.Message)
	assert.NotNil(t, badRequestResp.Details)
	details := *badRequestResp.Details
	assert.Equal(t, "must pass one search parameter", details["error"])
	assert.NotNil(t, details["reference_id"])

	// Verify no calls were made to the mock
	mockClient.AssertNotCalled(t, "SearchBooks")
}

func TestServer_GetApiV1Books_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		error         error
		expectedCode  string
		expectedMsg   string
		expectedError string
		expectedType  any
	}{
		{
			name:          "Bad Request Error",
			error:         openlibrary.ErrBadRequest,
			expectedCode:  "BAD_REQUEST",
			expectedMsg:   "search books - bad request",
			expectedError: openlibrary.ErrBadRequest.Error(),
			expectedType:  betterreads.GetApiV1Books400JSONResponse{},
		},
		{
			name:          "Not Found Error",
			error:         openlibrary.ErrNotFound,
			expectedCode:  "NOT_FOUND",
			expectedMsg:   "search books - not found",
			expectedError: openlibrary.ErrNotFound.Error(),
			expectedType:  betterreads.GetApiV1Books404JSONResponse{},
		},
		{
			name:          "Internal Server Error",
			error:         openlibrary.ErrInternalServer,
			expectedCode:  "INTERNAL_SERVER_ERROR",
			expectedMsg:   "search books - internal server error",
			expectedError: openlibrary.ErrInternalServer.Error(),
			expectedType:  betterreads.GetApiV1Books500JSONResponse{},
		},
		{
			name:          "Unknown Error",
			error:         errors.New("unknown error"),
			expectedCode:  "UNKNOWN",
			expectedMsg:   "search books - unknown",
			expectedError: "unknown error",
			expectedType:  betterreads.GetApiV1Books500JSONResponse{},
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
			queryPtr := testQuery
			resp, err := server.GetApiV1Books(context.Background(), betterreads.GetApiV1BooksRequestObject{
				Params: betterreads.GetApiV1BooksParams{
					Query: &queryPtr,
				},
			})

			// Assert
			assert.NoError(t, err)

			// Check response type and common fields
			switch actualResp := resp.(type) {
			case betterreads.GetApiV1Books400JSONResponse:
				assert.IsType(t, betterreads.GetApiV1Books400JSONResponse{}, tt.expectedType)
				assert.Equal(t, tt.expectedCode, actualResp.Code)
				assert.Equal(t, tt.expectedMsg, actualResp.Message)
				assert.NotNil(t, actualResp.Details)
				details := *actualResp.Details
				assert.Equal(t, tt.expectedError, details["error"])
				assert.NotNil(t, details["reference_id"])
			case betterreads.GetApiV1Books404JSONResponse:
				assert.IsType(t, betterreads.GetApiV1Books404JSONResponse{}, tt.expectedType)
				assert.Equal(t, tt.expectedCode, actualResp.Code)
				assert.Equal(t, tt.expectedMsg, actualResp.Message)
				assert.NotNil(t, actualResp.Details)
				details := *actualResp.Details
				assert.Equal(t, tt.expectedError, details["error"])
				assert.NotNil(t, details["reference_id"])
			case betterreads.GetApiV1Books500JSONResponse:
				assert.IsType(t, betterreads.GetApiV1Books500JSONResponse{}, tt.expectedType)
				assert.Equal(t, tt.expectedCode, actualResp.Code)
				assert.Equal(t, tt.expectedMsg, actualResp.Message)
				assert.NotNil(t, actualResp.Details)
				details := *actualResp.Details
				assert.Equal(t, tt.expectedError, details["error"])
				assert.NotNil(t, details["reference_id"])
			default:
				t.Fatalf("Unexpected response type: %T", resp)
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestServer_GetApiV1Books_EmptyResults(t *testing.T) {
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
	queryPtr := testQuery
	resp, err := server.GetApiV1Books(context.Background(), betterreads.GetApiV1BooksRequestObject{
		Params: betterreads.GetApiV1BooksParams{
			Query: &queryPtr,
		},
	})

	// Assert
	assert.NoError(t, err)

	// Type assertion to get the specific response type
	successResp, ok := resp.(betterreads.GetApiV1Books200JSONResponse)
	assert.True(t, ok, "Expected a 200 response")

	// Verify response content
	assert.Empty(t, successResp.Books)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestServer_GetApiV1Books_MultipleBooks(t *testing.T) {
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
				Source:          "open_library",
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
				Source:          "open_library",
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
				Source:          "open_library",
			},
		},
	}

	// Expectations
	mockClient.On("SearchBooks", mock.Anything, testQuery).Return(mockResponse, nil)

	// Execute
	queryPtr := testQuery
	resp, err := server.GetApiV1Books(context.Background(), betterreads.GetApiV1BooksRequestObject{
		Params: betterreads.GetApiV1BooksParams{
			Query: &queryPtr,
		},
	})

	// Assert
	assert.NoError(t, err)

	// Type assertion to get the specific response type
	successResp, ok := resp.(betterreads.GetApiV1Books200JSONResponse)
	assert.True(t, ok, "Expected a 200 response")

	// Verify response content
	assert.Len(t, successResp.Books, 3, "Expected 3 books in the response")

	// Verify first book
	assert.Equal(t, "OL456M", successResp.Books[0].Id)
	assert.Equal(t, "Author One", successResp.Books[0].AuthorName)
	assert.Equal(t, "OL123A", successResp.Books[0].AuthorId)
	assert.Equal(t, "First Book", successResp.Books[0].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL456M-L.jpg", successResp.Books[0].BookImage)
	assert.Equal(t, "1234567890", successResp.Books[0].Isbn)
	assert.Equal(t, float32(4.5), successResp.Books[0].RatingAverage)
	assert.Equal(t, 100, successResp.Books[0].RatingCount)
	assert.Equal(t, 2020, successResp.Books[0].PublishedYear)
	assert.Equal(t, "open_library", successResp.Books[0].Source)

	// Verify second book
	assert.Equal(t, "OL101M", successResp.Books[1].Id)
	assert.Equal(t, "Author Two", successResp.Books[1].AuthorName)
	assert.Equal(t, "OL789A", successResp.Books[1].AuthorId)
	assert.Equal(t, "Second Book", successResp.Books[1].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL101M-L.jpg", successResp.Books[1].BookImage)
	assert.Equal(t, "0987654321", successResp.Books[1].Isbn)
	assert.Equal(t, float32(3.8), successResp.Books[1].RatingAverage)
	assert.Equal(t, 75, successResp.Books[1].RatingCount)
	assert.Equal(t, 2018, successResp.Books[1].PublishedYear)
	assert.Equal(t, "open_library", successResp.Books[1].Source)

	// Verify third book
	assert.Equal(t, "OL202M", successResp.Books[2].Id)
	assert.Equal(t, "Author Three", successResp.Books[2].AuthorName)
	assert.Equal(t, "OL456A", successResp.Books[2].AuthorId)
	assert.Equal(t, "Third Book", successResp.Books[2].Title)
	assert.Equal(t, "https://covers.openlibrary.org/b/olid/OL202M-L.jpg", successResp.Books[2].BookImage)
	assert.Equal(t, "5678901234", successResp.Books[2].Isbn)
	assert.Equal(t, float32(4.2), successResp.Books[2].RatingAverage)
	assert.Equal(t, 50, successResp.Books[2].RatingCount)
	assert.Equal(t, 2022, successResp.Books[2].PublishedYear)
	assert.Equal(t, "open_library", successResp.Books[2].Source)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}
