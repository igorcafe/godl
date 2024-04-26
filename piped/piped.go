package piped

import (
	"context"
	"net/http"
)

type Service interface {
	FindStreams(ctx context.Context, videoID string) (StreamsResponse, error)
}

type service struct {
	baseURL string
	http    *http.Client
}

var _ Service = service{}

func NewService(baseURL string, httpClient *http.Client) Service {
	return service{
		baseURL,
		httpClient,
	}
}

type Instance struct {
	URL string
}

type InstanceService interface {
	List(ctx context.Context) ([]Instance, error)
	GetTopN(ctx context.Context, n int, instances []Instance) ([]Instance, error)
}

type instanceService struct {
	http *http.Client
}

var _ InstanceService = instanceService{}
