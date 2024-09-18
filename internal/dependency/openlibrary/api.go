package openlibrary

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	iLogger "github.com/celestialdragonfly/betterreads-platform/internal/package/log"
)

type API struct {
	Log *iLogger.Logger
}

func NewAPI(log *iLogger.Logger) *API {
	return &API{
		Log: log,
	}
}

const (
	queryString   = "?q="
	bookSearchURL = "https://openlibrary.org/search.json"
	authSearchURL = ""
)

var (
	ErrorRead            = errors.New("unable to retrieve books")
	ErrorParse           = errors.New("unable to parse books")
	ErrorUnmarshal       = errors.New("unable to unmarshal books")
	DefaultBookFieldMask = []string{"cover_i", "title", "author_name", "key", "author_key", "ratings_average", "ratings_count", "isbn", "cover_edition_key"}
)

type SearchBookRequest struct {
	Query     string
	FieldMask []string
	Page      int
}

type SearchBooksResponse struct {
	Books []struct {
		AuthorKey       []string
		AuthorName      []string
		CoverEditionKey string
		CoverImage      string
		ISBN            string
		Key             string
		Title           string
		Description     string
		RatingAverage   float64
		RatingCount     int
	}
}

func (a API) SearchBooks(query string, fieldMask []string, page int) (*SearchBooksResponse, error) {
	url := buildBaseQuery(bookSearchURL, query)
	if len(fieldMask) > 0 {
		url = appendFieldMask(url, fieldMask)
	}
	if page > 1 {
		url = appendPagination(url, page)
	}
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		a.Log.Error("SearchBooks: unable to make http GET request to the OpenLibrary API", iLogger.Fields{"error": err.Error()})
		return nil, ErrorRead
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.Log.Error("SearchBooks: unable to read the response body from the OpenLibrary API", iLogger.Fields{"error": err.Error()})
		return nil, ErrorParse
	}
	fmt.Println(body)

	type OpenAPIResponse struct {
		NumFound      int  `json:"numFound"`
		Start         int  `json:"start"`
		NumFoundExact bool `json:"numFoundExact"`
		Docs          []struct {
			AuthorKey       []string `json:"author_key,omitempty"`
			AuthorName      []string `json:"author_name,omitempty"`
			CoverEditionKey string   `json:"cover_edition_key,omitempty"`
			CoverI          int      `json:"cover_i,omitempty"`
			Isbn            []string `json:"isbn,omitempty"`
			Key             string   `json:"key"`
			Title           string   `json:"title"`
			RatingsAverage  float64  `json:"ratings_average,omitempty"`
			RatingCount     int      `json:"rating_count,omitempty"`
		} `json:"docs"`
	}

	var oAPIResponse OpenAPIResponse

	if err = json.Unmarshal(body, &oAPIResponse); err != nil {
		a.Log.Error("SearchBooks: unable to unmarshal the response body from the OpenLibrary API", iLogger.Fields{"error": err.Error()})
		return nil, ErrorUnmarshal
	}

	return &books, nil
}

func GetBook() {}

func GetAuthor() {

}

func GetBookThumbnail() {

}

func GetAuthorThumbnail() {

}

func buildBaseQuery(url, query string) string {
	return url + queryString + strings.ReplaceAll(query, " ", "+")
}

func appendFieldMask(url string, fieldMask []string) string {
	url = url + "&fields="
	for i, mask := range fieldMask {
		url = url + mask
		if i < len(fieldMask)-1 {
			url = url + ","
		}
	}
	return url
}

func appendPagination(url string, page int) string {
	return fmt.Sprintf("%s&page=%d", url, page)
}
