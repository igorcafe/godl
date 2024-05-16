package media

import (
	"context"
	"os/exec"
)

type Service interface {
	JoinTracks(ctx context.Context, dest string, sources ...string) error
}

type service struct{}

func NewFFmpegService() Service {
	return service{}
}

func (s service) JoinTracks(ctx context.Context, dest string, sources ...string) error {
	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel",
		"error",
	}

	for _, src := range sources {
		args = append(args, "-i", src)
	}
	args = append(args, "-c", "copy", dest)

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	_, err := cmd.CombinedOutput()

	return err
}
