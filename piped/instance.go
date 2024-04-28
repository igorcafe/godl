package piped

import (
	"context"
	"net/http"
)

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

func NewInstanceService(http *http.Client) InstanceService {
	return instanceService{
		http,
	}
}
