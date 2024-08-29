package contracts

import "github.com/google/uuid"

type GoogleBook struct {
	ID             string   `json:"id"`
	Etag           string   `json:"etag"`
	SelfLink       string   `json:"self_link"`
	Title          string   `json:"title"`
	Subtitle       string   `json:"subtitle"`
	Authors        []string `json:"authors"`
	Categories     []string `json:"categories"`
	AverageRating  int      `json:"average_rating"`
	RatingsCount   int      `json:"ratings_count"`
	MaturityRating string   `json:"maturity_rating"`
	SmallThumbnail string   `json:"small_thumbnail"`
	Thumbnail      string   `json:"thumbnail"`
	PreviewLink    string   `json:"preview_link"`
	InfoLink       string   `json:"info_link"`
}

type GetBooksRequest struct{}

type GetBooksResponse struct {
	Books []GoogleBook `json:"books"`
}

type GetBooksForUserRequest struct{}
type GetBooksForUserResponse struct {
}

type HealthCheckRequest struct{}

type HealthCheckResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environement"`
	Version     string `json:"version"`
}

type ErrorResponse struct {
	Message     string `json:"msg"`
	ReferenceID string `json:"reference_id"`
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Message:     message,
		ReferenceID: uuid.New().String(),
	}
}
