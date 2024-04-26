package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	urlpkg "net/url"

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
		log.Fatal("no videoID")
	}

	instanceSvc := piped.NewInstanceService(http.DefaultClient)
	instances, err := instanceSvc.List(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	top, err := instanceSvc.GetTopN(context.Background(), 1, instances)
	if err != nil {
		log.Fatal(err)
	}

	pipedSvc := piped.NewService(top[0].URL, http.DefaultClient)

	streams, err := pipedSvc.FindStreams(context.Background(), videoID)
	if err != nil {
		log.Fatal(err)
	}

	videoStream := streams.VideoStreams[len(streams.VideoStreams)-1]
	audioStream := streams.AudioStreams[len(streams.AudioStreams)-1]

	dlSvc := download.NewService(http.DefaultClient)

	errs := make(chan error)

	videoPath := "video.mp4"
	go func() {
		err := dlSvc.DownloadStream(context.Background(), videoStream.URL, videoPath, func(elapsed, total int64) {
			fmt.Printf("video: %d KB / %d KB\n", elapsed/1024, total/1024)
		})
		errs <- err
	}()

	audioPath := "audio.mp3"
	go func() {
		err := dlSvc.DownloadStream(context.Background(), audioStream.URL, audioPath, func(elapsed, total int64) {
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
