package openlibrary

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	library "github.com/celestialdragonfly/betterreads/internal/openlibrary/contracts"
)

// MockHTTPClient is a custom HTTP client for testing purposes.
type MockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do implements the HttpRequestDoer interface.
func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// TestNewClient tests the NewClient function.
func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		host    string
		wantErr bool
	}{
		{
			name:    "Valid host",
			host:    "https://openlibrary.org",
			wantErr: false,
		},
		{
			name:    "Invalid host",
			host:    "://invalid-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() expected non-nil client but got nil")
			}
		})
	}
}

// TestSearchBooks tests the SearchBooks function.
//
//nolint:gocognit // unit test
func TestSearchBooks(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		query          string
		title          *string
		author         *string
		subject        *string
		mockResponse   *http.Response
		mockErr        error
		wantErr        bool
		expectedErrMsg string
		expectedBooks  int
	}{
		{
			name:  "Successful search with multiple books",
			query: "harry potter",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"numFound": 2,
					"start": 0,
					"numFoundExact": true,
					"docs": [
						{
							"key": "/works/OL82563W",
							"title": "Harry Potter and the Philosopher's Stone",
							"cover_edition_key": "OL22856696M",
							"author_name": ["J. K. Rowling"],
							"author_key": ["OL23919A"],
							"isbn": ["9780747532743"],
							"publish_year": [1997],
							"ratings_average": 4.5,
							"ratings_count": 1000
						},
						{
							"key": "/works/OL82564W",
							"title": "Harry Potter and the Chamber of Secrets",
							"cover_edition_key": "OL22856697M",
							"author_name": ["J. K. Rowling"],
							"author_key": ["OL23919A"],
							"isbn": ["9780747538493"],
							"publish_year": [1998],
							"ratings_average": 4.4,
							"ratings_count": 900
						}
					],
					"q": "harry potter",
					"offset": null,
					"documentation_url": "https://openlibrary.org/dev/docs/api/search"
				}`)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			mockErr:       nil,
			wantErr:       false,
			expectedBooks: 2,
		},
		{
			name:  "Empty response",
			query: "harry potter",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"numFound": 0,
					"start": 0,
					"numFoundExact": true,
					"docs": [],
					"q": "nonexistentbook123456789",
					"offset": null,
					"documentation_url": "https://openlibrary.org/dev/docs/api/search"
				}`)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			mockErr:       nil,
			wantErr:       false,
			expectedBooks: 0,
		},
		{
			name:           "Server error",
			mockResponse:   nil,
			mockErr:        errors.New("connection refused"),
			wantErr:        true,
			expectedErrMsg: "SearchBooks: internal server error",
			expectedBooks:  0,
		},
		{
			name:  "Invalid status code",
			query: "harry potter",
			mockResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`{"error": "bad request"}`)),
			},
			mockErr:        nil,
			wantErr:        true,
			expectedErrMsg: "SearchBooks: bad request",
			expectedBooks:  0,
		},
		{
			name:  "Malformed response",
			query: "harry potter",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"invalid json"`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
			mockErr:        nil,
			wantErr:        true,
			expectedErrMsg: "SearchBooks: not found",
			expectedBooks:  0,
		},
		{
			name:  "Response with books but no cover edition key",
			query: "harry potter",
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"numFound": 1,
					"start": 0,
					"numFoundExact": true,
					"docs": [
						{
							"key": "/works/OL82563W",
							"title": "Harry Potter and the Philosopher's Stone",
							"cover_edition_key": "",
							"author_name": ["J. K. Rowling"],
							"author_key": ["OL23919A"],
							"isbn": ["9780747532743"],
							"publish_year": [1997],
							"ratings_average": 4.5,
							"ratings_count": 1000
						}
					],
					"q": "harry potter",
					"offset": null,
					"documentation_url": "https://openlibrary.org/dev/docs/api/search"
				}`)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			mockErr:       nil,
			wantErr:       false,
			expectedBooks: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP client
			mockClient := &MockHTTPClient{
				DoFunc: func(*http.Request) (*http.Response, error) {
					return tt.mockResponse, tt.mockErr
				},
			}

			// Create a library client with the mock HTTP client
			libraryClient, err := library.NewClient("https://openlibrary.org", library.WithHTTPClient(mockClient))
			if err != nil {
				t.Fatalf("Failed to create library client: %v", err)
			}

			// Create our client with the mocked library client
			client := &Client{
				Client: libraryClient,
			}

			// Call the function being tested
			result, err := client.SearchBooks(ctx, tt.query, tt.title, tt.author, tt.subject)

			// Check error expectations
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchBooks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.expectedErrMsg != "" {
				if !strings.Contains(err.Error(), tt.expectedErrMsg) {
					t.Errorf("SearchBooks() error message = %v, expected to contain %v", err.Error(), tt.expectedErrMsg)
				}
				return
			}

			if result == nil {
				t.Fatal("SearchBooks() returned nil result")
			}

			if len(result.Books) != tt.expectedBooks {
				t.Errorf("SearchBooks() returned %d books, expected %d", len(result.Books), tt.expectedBooks)
			}

			// Check a few attributes of the first book if we expect books
			if tt.expectedBooks > 0 {
				book := result.Books[0]
				if book.Title == "" {
					t.Error("SearchBooks() returned a book with empty title")
				}
				if book.CoverEditionKey == "" {
					t.Error("SearchBooks() returned a book with empty cover edition key")
				}
				if book.CoverImage == "" {
					t.Error("SearchBooks() returned a book with empty cover image")
				}
				if book.Source != "open_library" {
					t.Errorf("SearchBooks() returned a book with source = %s, expected 'open_library'", book.Source)
				}
			}
		})
	}
}

// TestGetFirstValue tests the getFirstValue function.
func TestGetFirstValue(t *testing.T) {
	// Test with strings
	t.Run("String slice with values", func(t *testing.T) {
		slice := []string{"first", "second", "third"}
		result := getFirstValue(slice)
		expected := "first"
		if result != expected {
			t.Errorf("getFirstValue() = %v, want %v", result, expected)
		}
	})

	t.Run("Empty string slice", func(t *testing.T) {
		slice := []string{}
		result := getFirstValue(slice)
		expected := ""
		if result != expected {
			t.Errorf("getFirstValue() = %v, want %v", result, expected)
		}
	})

	// Test with integers
	t.Run("Integer slice with values", func(t *testing.T) {
		slice := []int{1, 2, 3}
		result := getFirstValue(slice)
		expected := 1
		if result != expected {
			t.Errorf("getFirstValue() = %v, want %v", result, expected)
		}
	})

	t.Run("Empty integer slice", func(t *testing.T) {
		slice := []int{}
		result := getFirstValue(slice)
		expected := 0
		if result != expected {
			t.Errorf("getFirstValue() = %v, want %v", result, expected)
		}
	})
}

// TestSearchBooksIntegration tests the full SearchBooks functionality with a mocked response.
//
//nolint:gocognit // unit test
func TestSearchBooksIntegration(t *testing.T) {
	ctx := context.Background()
	query := "harry potter"

	// Mock the full response flow
	mockClient := &MockHTTPClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			// Verify the request
			if !strings.Contains(req.URL.String(), "search.json") {
				t.Errorf("Expected URL to contain 'search.json', got %s", req.URL.String())
			}
			if req.Method != http.MethodGet {
				t.Errorf("Expected method GET, got %s", req.Method)
			}
			if !strings.Contains(req.URL.RawQuery, "q=harry+potter") {
				t.Errorf("Expected query to contain 'q=harry+potter', got %s", req.URL.RawQuery)
			}
			if req.Header.Get("User-Agent") == "" {
				t.Error("Expected User-Agent header to be set")
			}

			// Return a successful response
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"numFound": 2,
					"start": 0,
					"numFoundExact": true,
					"docs": [
						{
							"key": "/works/OL82563W",
							"title": "Harry Potter and the Philosopher's Stone",
							"cover_edition_key": "OL22856696M",
							"author_name": ["J. K. Rowling"],
							"author_key": ["OL23919A"],
							"isbn": ["9780747532743"],
							"publish_year": [1997],
							"ratings_average": 4.5,
							"ratings_count": 1000
						},
						{
							"key": "/works/OL82564W",
							"title": "Harry Potter and the Chamber of Secrets",
							"cover_edition_key": "OL22856697M",
							"author_name": ["J. K. Rowling"],
							"author_key": ["OL23919A"],
							"isbn": ["9780747538493"],
							"publish_year": [1998],
							"ratings_average": 4.4,
							"ratings_count": 900
						}
					],
					"q": "harry potter",
					"offset": null,
					"documentation_url": "https://openlibrary.org/dev/docs/api/search"
				}`)),
				Header: http.Header{"Content-Type": []string{"application/json"}},
			}, nil
		},
	}

	// Create a library client with the mock HTTP client
	libraryClient, err := library.NewClient("https://openlibrary.org", library.WithHTTPClient(mockClient))
	if err != nil {
		t.Fatalf("Failed to create library client: %v", err)
	}

	// Create our client with the mocked library client
	client := &Client{
		Client: libraryClient,
	}

	// Call the function being tested
	result, err := client.SearchBooks(ctx, query, nil, nil, nil)
	if err != nil {
		t.Fatalf("SearchBooks() error = %v", err)
	}

	// Verify the response
	if len(result.Books) != 2 {
		t.Errorf("Expected 2 books, got %d", len(result.Books))
	}

	// Check the first book's details
	expectedFirstBook := Book{
		AuthorKey:       "OL23919A",
		AuthorName:      "J. K. Rowling",
		CoverEditionKey: "OL22856696M",
		CoverImage:      "https://covers.openlibrary.org/b/olid/OL22856696M-L.jpg",
		ISBN:            "9780747532743",
		Title:           "Harry Potter and the Philosopher's Stone",
		RatingAverage:   4.5,
		RatingCount:     1000,
		PublishYear:     1997,
		Source:          "open_library",
	}

	book := result.Books[0]

	// Check individual fields instead of using DeepEqual to make test less brittle
	if book.Title != expectedFirstBook.Title {
		t.Errorf("Book title mismatch. Expected: %s, Got: %s", expectedFirstBook.Title, book.Title)
	}
	if book.AuthorName != expectedFirstBook.AuthorName {
		t.Errorf("Book author mismatch. Expected: %s, Got: %s", expectedFirstBook.AuthorName, book.AuthorName)
	}
	if book.AuthorKey != expectedFirstBook.AuthorKey {
		t.Errorf("Book author key mismatch. Expected: %s, Got: %s", expectedFirstBook.AuthorKey, book.AuthorKey)
	}
	if book.CoverEditionKey != expectedFirstBook.CoverEditionKey {
		t.Errorf("Book cover edition key mismatch. Expected: %s, Got: %s", expectedFirstBook.CoverEditionKey, book.CoverEditionKey)
	}
	if book.CoverImage != expectedFirstBook.CoverImage {
		t.Errorf("Book cover image mismatch. Expected: %s, Got: %s", expectedFirstBook.CoverImage, book.CoverImage)
	}
	if book.ISBN != expectedFirstBook.ISBN {
		t.Errorf("Book ISBN mismatch. Expected: %s, Got: %s", expectedFirstBook.ISBN, book.ISBN)
	}
	if book.PublishYear != expectedFirstBook.PublishYear {
		t.Errorf("Book publish year mismatch. Expected: %d, Got: %d", expectedFirstBook.PublishYear, book.PublishYear)
	}
	if book.RatingAverage != expectedFirstBook.RatingAverage {
		t.Errorf("Book rating average mismatch. Expected: %f, Got: %f", expectedFirstBook.RatingAverage, book.RatingAverage)
	}
	if book.RatingCount != expectedFirstBook.RatingCount {
		t.Errorf("Book rating count mismatch. Expected: %d, Got: %d", expectedFirstBook.RatingCount, book.RatingCount)
	}
	if book.Source != expectedFirstBook.Source {
		t.Errorf("Book source mismatch. Expected: %s, Got: %s", expectedFirstBook.Source, book.Source)
	}
}
