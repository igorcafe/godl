package piped

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
)

func (s instanceService) GetTopN(ctx context.Context, n int, instances []Instance) ([]Instance, error) {
	if n > len(instances) {
		n = len(instances)
	}

	ch := make(chan Instance, n)
	ctx, cancel := context.WithCancel(ctx)

	for _, inst := range instances {
		inst := inst
		piped := NewService(log.New(io.Discard, "", 0), inst.URL, s.http)
		go func() {
			res, err := piped.FindStreams(ctx, "oRdxUFDoQe0")
			if errors.Is(err, context.Canceled) {
				return
			}
			if err != nil {
				// log.Print("failed to interact with instance: ", err)
				return
			}

			if len(res.VideoStreams) == 0 {
				return
			}

			url := res.VideoStreams[len(res.VideoStreams)-1].URL

			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return
			}

			resp, err := s.http.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				return
			}

			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				return
			}

			ch <- inst
		}()
	}

	result := make([]Instance, n)
	for i := 0; i < n; i++ {
		result[i] = <-ch
	}
	cancel()

	return result, nil
}
