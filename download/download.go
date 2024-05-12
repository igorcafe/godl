package download

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/igoracmelo/godl/log"
	"github.com/igoracmelo/godl/media"
	"github.com/igoracmelo/godl/piped"
)

type Service interface {
	DownloadStream(ctx context.Context, url string, destPath string, onProgress func(int64, int64)) error
	DownloadFromPiped(ctx context.Context, params DownloadFromPipedParams) error
}

type DownloadFromPipedParams struct {
	Title        string
	AudioStream  *piped.AudioStream
	VideoStream  *piped.VideoStream
	MediaService media.Service
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

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tempPath := "partial." + destPath
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
	}()

	elapsed := int64(0)
	total := resp.ContentLength

	// Can't use io.Copy since I want to call onProgress
	buf := make([]byte, 132*1024)
	for {
		select {
		default:
		case <-ctx.Done():
			return context.Canceled
		}

		nr, rerr := resp.Body.Read(buf)
		if nr == 0 && rerr == io.EOF {
			break
		}

		nw, werr := tempFile.Write(buf[:nr])
		if werr != nil {
			return werr
		}

		elapsed += int64(nw)
		onProgress(elapsed, total)

		if nw != nr {
			return io.ErrShortWrite
		}

		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return rerr
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

func (s service) DownloadFromPiped(ctx context.Context, params DownloadFromPipedParams) error {
	errs := make(chan error)
	downloadCount := 0

	videoPath := "video." + params.Title + ".mp4"
	if params.VideoStream != nil {
		downloadCount++
		go func() {
			now := time.Now()
			log.Infof("video: download started")
			err := s.DownloadStream(ctx, params.VideoStream.URL, videoPath, func(elapsed, total int64) {
				log.Infof("video: %d KB / %d KB", elapsed/1024, total/1024)
			})
			if err == nil {
				log.Infof("video: finished downloading in %.1f seconds", time.Since(now).Seconds())
			}
			errs <- err
		}()
	}

	audioPath := "audio." + params.Title + ".mp3"
	if params.AudioStream != nil {
		downloadCount++
		go func() {
			now := time.Now()
			log.Infof("audio: download started")
			err := s.DownloadStream(ctx, params.AudioStream.URL, audioPath, func(elapsed, total int64) {
				log.Infof("audio: %d KB / %d KB", elapsed/1024, total/1024)
			})
			if err == nil {
				log.Infof("audio: finished downloading in %.1f seconds", time.Since(now).Seconds())
			}
			errs <- err
		}()
	}

	for i := 0; i < downloadCount; i++ {
		err := <-errs
		if err != nil {
			return err
		}
	}

	if downloadCount == 0 {
		return errors.New("nothing to download")
	}

	finalPath := params.Title + ".mp4"
	if downloadCount > 1 {
		now := time.Now()
		log.Infof("joining audio and video tracks")

		err := params.MediaService.JoinTracks(ctx, finalPath, audioPath, videoPath)
		if err != nil {
			log.Fatalf("failed to join tracks: %v", err)
		}
		log.Infof("finished joining tracks in %.1f seconds", time.Since(now).Seconds())

		err = errors.Join(os.Remove(videoPath), os.Remove(audioPath))
		if err != nil {
			log.Fatalf("failed to join tracks: %v", err)
		}
	}

	return nil
}
