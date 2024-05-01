package piped

import (
	"context"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	t.Parallel()
	s := instanceService{
		http.DefaultClient,
	}

	list, err := s.List(context.Background())
	require.NoError(t, err)

	list, err = s.GetTopN(context.Background(), 5, list)
	require.NoError(t, err)

	log.Printf(">>>> list: %#+v\n", list)
}
