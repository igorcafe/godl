package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	urlpkg "net/url"
	"os"
	"os/signal"
	"strconv"

	"github.com/igoracmelo/godl/bufioctx"
	"github.com/igoracmelo/godl/download"
	"github.com/igoracmelo/godl/piped"
)

func main() {
	flag.Parse()

	stderr := log.New(os.Stderr, "", 0)

	url, err := urlpkg.Parse(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	videoID := url.Query().Get("v")
	if videoID == "" {
		stderr.Panic("no videoID")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	instanceSvc := piped.NewInstanceService(http.DefaultClient)
	instances, err := instanceSvc.List(ctx)
	if err != nil {
		stderr.Panic(err)
	}

	top, err := instanceSvc.GetTopN(ctx, 1, instances)
	if err != nil {
		stderr.Panic(err)
	}

	fmt.Println("using instance", top[0].URL)
	pipedSvc := piped.NewService(top[0].URL, http.DefaultClient)

	streams, err := pipedSvc.FindStreams(ctx, videoID)
	if err != nil {
		stderr.Panic(err)
	}

	lines := bufioctx.NewScanner(os.Stdin)

	var videoOption int
	fmt.Println("0) No video")
	for i, stream := range streams.VideoStreams {
		fmt.Printf("%d) %s\n", i+1, stream.Quality)
	}
	fmt.Print("Choose video option (0 to skip): ")
	lines.Scan(ctx)
	videoOption, err = strconv.Atoi(lines.Text())
	if err != nil {
		stderr.Panic("invalid index")
	}
	if videoOption < 0 || videoOption > len(streams.VideoStreams) {
		stderr.Panic("invalid index")
	}
	fmt.Println()

	var audioOption int
	fmt.Println("0) No audio")
	for i, stream := range streams.AudioStreams {
		fmt.Printf("%d) %s - %s - %s\n", i+1, stream.Quality, stream.Codec, stream.MIMEType)
	}
	fmt.Print("Choose audio option (0 to skip): ")
	lines.Scan(ctx)
	audioOption, err = strconv.Atoi(lines.Text())
	if err != nil {
		stderr.Panic("invalid index")
	}
	if audioOption < 0 || audioOption > len(streams.AudioStreams) {
		stderr.Panic("invalid index")
	}
	fmt.Println()

	dlSvc := download.NewService(http.DefaultClient)
	errs := make(chan error)
	downloadCount := 0

	if videoOption != 0 {
		downloadCount++
		go func() {
			videoPath := "video.mp4"
			videoStream := streams.VideoStreams[videoOption-1]
			err := dlSvc.DownloadStream(ctx, videoStream.URL, videoPath, func(elapsed, total int64) {
				stderr.Printf("video: %d KB / %d KB\n", elapsed/1024, total/1024)
			})
			errs <- err
		}()
	}

	if audioOption != 0 {
		downloadCount++
		go func() {
			audioPath := "audio.mp3"
			audioStream := streams.AudioStreams[len(streams.AudioStreams)-1]
			err := dlSvc.DownloadStream(ctx, audioStream.URL, audioPath, func(elapsed, total int64) {
				stderr.Printf("audio: %d KB / %d KB\n", elapsed/1024, total/1024)
			})
			errs <- err
		}()
	}

	for i := 0; i < downloadCount; i++ {
		err = <-errs
		if err != nil {
			stderr.Print(err)
		}
	}

	if downloadCount == 0 {
		fmt.Println("nothing to download")
		os.Exit(1)
	}
}
