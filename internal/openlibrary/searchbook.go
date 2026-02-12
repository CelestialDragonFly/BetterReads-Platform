package openlibrary

import (
	"context"
	"fmt"
	"strings"

	"github.com/celestialdragonfly/betterreads/internal/logger"
	library "github.com/celestialdragonfly/betterreads/internal/openlibrary/contracts"
)

type Book struct {
	AuthorKey       string
	AuthorName      string
	CoverEditionKey string
	CoverImage      string
	ISBN            string
	Title           string
	RatingAverage   float32
	RatingCount     int
	PublishYear     int
	Source          string
}
type SearchBooksResponse struct {
	Books []Book
}

// SearchBooks queries the open library API for books matching the given query string.
// It fetches book details such as author, edition key, ISBN, title, ratings, and publication year.
// The function returns a SearchBooksResponse containing a list of books or an error if the search fails.
//
// Parameters:
// - ctx: The context for the request, allowing for cancellation and timeout control.
// - query: The search term used to find books in the library database.
//
// Returns:
// - *SearchBooksResponse: A response containing a list of books matching the search criteria.
// - error: An error if the request fails, response is empty, or an issue occurs during processing.
func (c *Client) SearchBooks(ctx context.Context, query string, title, author, subject *string) (*SearchBooksResponse, error) {
	var (
		defaultBookFieldMask = strings.Join(
			[]string{
				"author_key",
				"author_name",
				"cover_edition_key", // book id
				"isbn",
				"title",
				"ratings_average",
				"ratings_count",
				"publish_year",
			}, ",")
		resultLimit     = 15
		defaultLanguage = "en"
	)
	resp, err := c.Client.SearchBooks(ctx, &library.SearchBooksParams{
		UserAgent: "BetterReads robertjhird@gmail.com",
		Q:         query,
		Title:     title,
		Author:    author,
		Subject:   subject,
		Limit:     &resultLimit,
		Fields:    &defaultBookFieldMask,
		Lang:      &defaultLanguage,
	})
	if err != nil {
		return nil, fmt.Errorf("SearchBooks: %w - search error, error: %w", ErrInternalServer, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warn("unable to close response body", map[string]interface{}{"error": closeErr})
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("SearchBooks: %w, library returned non 200 status", ErrBadRequest)
	}

	searchResponse, err := library.ParseSearchBooksResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("SearchBooks: %w - failed to parse response: %w", ErrNotFound, err)
	}

	books := searchResponse.JSON200.Docs

	var bookResponse []Book
	for _, book := range books {
		olid := book.CoverEditionKey
		if olid == "" {
			continue
		}
		bookResponse = append(bookResponse, Book{
			AuthorKey:  getFirstValue(book.AuthorKey),
			AuthorName: getFirstValue(book.AuthorName),
			// TODO maybe give author image
			CoverEditionKey: olid,
			CoverImage:      fmt.Sprintf("https://covers.openlibrary.org/b/olid/%s-L.jpg", olid),
			ISBN:            getFirstValue(book.Isbn),
			Title:           book.Title,
			RatingAverage:   book.RatingsAverage,
			RatingCount:     book.RatingsCount,
			PublishYear:     getFirstValue(book.PublishYear),
			Source:          "open_library",
		})
	}

	// Return empty results instead of an error when no books have cover edition keys
	return &SearchBooksResponse{
		Books: bookResponse,
	}, nil
}
