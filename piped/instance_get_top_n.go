package piped

import (
	"context"
	"errors"
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
		go func() {
			req, err := http.NewRequestWithContext(ctx, "GET", inst.URL+"/streams/mtaQroi75M0", nil)
			if errors.Is(err, context.Canceled) {
				return
			}
			if err != nil {
				log.Print("failed to interact with instance: ", err)
				return
			}

			resp, err := s.http.Do(req)
			if errors.Is(err, context.Canceled) {
				return
			}
			if err != nil {
				log.Print("failed to interact with instance: ", err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Print("failed to interact with instance: ", resp.StatusCode)
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
