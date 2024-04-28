package bufioctx

import (
	"bufio"
	"context"
	"io"
)

type Scanner struct {
	s    *bufio.Scanner
	done chan struct{}
}

func NewScanner(r io.Reader) Scanner {
	return Scanner{
		s:    bufio.NewScanner(r),
		done: make(chan struct{}),
	}
}

func (s Scanner) Scan(ctx context.Context) bool {
	var ok bool
	go func() {
		ok = s.s.Scan()
		s.done <- struct{}{}
	}()

	select {
	case <-s.done:
		return ok
	case <-ctx.Done():
		return false
	}
}

func (s Scanner) Text() string {
	return s.s.Text()
}
