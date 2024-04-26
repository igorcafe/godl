package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"strings"

	"github.com/igoracmelo/godl/download"
	"github.com/igoracmelo/godl/piped"
)

func main() {
	flag.Parse()

	url, err := urlpkg.Parse(flag.Arg(0))
	if err != nil {
		panic(err)
	}

	fmt.Println(url.String())

	videoID := url.Query().Get("v")
	if videoID == "" {
		panic("no videoID")
	}

	pipedSvc := piped.NewService("https://api.piped.projectsegfau.lt", http.DefaultClient)

	streams, err := pipedSvc.FindStreams(context.Background(), videoID)
	if err != nil {
		panic(err)
	}

	videoStream := streams.VideoStreams[len(streams.VideoStreams)-1]
	audioStream := streams.AudioStreams[len(streams.AudioStreams)-1]

	dlService := download.NewService()

	errs := make(chan error)

	videoPath := "video." + strings.ToLower(videoStream.Format)
	go func() {
		err := dlService.DownloadStream(videoStream.URL, videoPath, func(elapsed, total int64) {
			fmt.Printf("video: %d KB / %d KB\n", elapsed/1024, total/1024)
		})
		errs <- err
	}()

	audioPath := "audio." + strings.ToLower(audioStream.Format)
	go func() {
		err := dlService.DownloadStream(videoStream.URL, audioPath, func(elapsed, total int64) {
			fmt.Printf("audio: %d KB / %d KB\n", elapsed/1024, total/1024)
		})
		errs <- err
	}()

	err = <-errs
	if err != nil {
		panic(err)
	}

	err = <-errs
	if err != nil {
		panic(err)
	}
}
