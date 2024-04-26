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
	List() ([]Instance, error)
	GetTopNInstances(n int) ([]Instance, error)
}

type instanceService struct {
	http *http.Client
}

var _ InstanceService = instanceService{}

func (s instanceService) List() ([]Instance, error) {
	return nil, nil
}

func (s instanceService) GetTopNInstances(n int) ([]Instance, error) {
	return nil, nil
}
