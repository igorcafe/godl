package piped

import (
	"context"
	"log"
	"net/http"
)

type Service interface {
	FindStreams(ctx context.Context, videoID string) (StreamsResponse, error)
}

type service struct {
	log     *log.Logger
	baseURL string
	http    *http.Client
}

var _ Service = service{}

func NewService(logger *log.Logger, baseURL string, httpClient *http.Client) Service {
	return service{
		logger,
		baseURL,
		httpClient,
	}
}
