package google

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	iLogger "github.com/celestialdragonfly/betterreads-platform/internal/package/log"
)

type BookAPI struct {
	URL string
	Key string
}

func NewAPI(apiKey string) *BookAPI {
	// flowers+inauthor:keyes&key="
	url := "https://www.googleapis.com/books/v1/volumes"

	return &BookAPI{
		URL: url,
		Key: apiKey,
	}
}

var (
	ErrorRead      = errors.New("unable to retrieve books")
	ErrorParse     = errors.New("unable to parse books")
	ErrorUnmarshal = errors.New("unable to unmarshal books")
)

func (api *BookAPI) GetBooks(log *iLogger.Logger, wildcard string) (*GoogleBooksResponse, error) {
	wildcard = strings.ReplaceAll(wildcard, " ", "+")
	url := fmt.Sprintf("%s?q=%s&orderBy=relevance&key=%s", api.URL, wildcard, api.Key)
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		log.Error("GetBooks: unable to make http GET request to the Google Books API", iLogger.Fields{"error": err.Error()})
		return nil, ErrorRead
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("GetBooks: unable to read the response body from the Google Books API", iLogger.Fields{"error": err.Error()})
		return nil, ErrorParse
	}

	var googleBooks GoogleBooksResponse
	if err = json.Unmarshal(body, &googleBooks); err != nil {
		log.Error("GetBooks: unable to unmarshal the response body from the Google Books API", iLogger.Fields{"error": err.Error()})
		return nil, err
	}
	return &googleBooks, nil
}
