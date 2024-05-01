package piped

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_InstanceService_List(t *testing.T) {
	t.Parallel()
	s := instanceService{
		http.DefaultClient,
	}
	inst, err := s.List(context.Background())
	require.NoError(t, err)
	require.Greater(t, len(inst), 10)
	// fmt.Println(inst)
}
