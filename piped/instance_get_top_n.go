package piped

import (
	"context"
	"errors"
	"io"
	"log"
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
			_, err := piped.FindStreams(ctx, "mtaQroi75M0")
			if errors.Is(err, context.Canceled) {
				return
			}
			if err != nil {
				// log.Print("failed to interact with instance: ", err)
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
