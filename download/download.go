package download

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
)

type Service interface {
	DownloadStream(ctx context.Context, url string, destPath string, onProgress func(int64, int64)) error
}

type service struct {
	http *http.Client
}

var _ Service = service{}

func NewService(httpClient *http.Client) Service {
	return service{
		httpClient,
	}
}

func (s service) DownloadStream(ctx context.Context, url string, destPath string, onProgress func(int64, int64)) error {
	if onProgress == nil {
		onProgress = func(i1, i2 int64) {}
	}

	resp, err := s.http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tempPath := "partial." + destPath
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer tempFile.Close()

	elapsed := int64(0)
	total := resp.ContentLength

	// Can't use io.Copy since I want to call onProgress
	buf := make([]byte, 132*1024)
	for {
		nr, rerr := resp.Body.Read(buf)
		nw, werr := tempFile.Write(buf[:nr])

		if errors.Is(rerr, io.EOF) {
			break
		}

		elapsed += int64(nw)
		onProgress(elapsed, total)

		if rerr != nil {
			return rerr
		}
		if werr != nil {
			return werr
		}
	}

	// make sure temp file is flushed to disk
	if err := tempFile.Sync(); err != nil {
		return err
	}

	// now that it is flushed, it can be considered completed
	if err := os.Rename(tempPath, destPath); err != nil {
		return err
	}

	return nil
}
