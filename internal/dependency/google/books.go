package google

import (
	"fmt"
	"io"
	"log"
	"net/http"
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

func (api *BookAPI) GetBooks(req string) {
	url := fmt.Sprintf("%s?q=intitle:%s+inauthor:%s+subject:%s&key=%s", api.URL, req, req, req, api.Key)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making GET request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Println(string(body))

}
