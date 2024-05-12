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
	"time"

	"github.com/igoracmelo/godl/bufioctx"
	"github.com/igoracmelo/godl/download"
	"github.com/igoracmelo/godl/media"
	"github.com/igoracmelo/godl/piped"
)

func main() {
	stderr := log.New(os.Stderr, "", 0)
	flag.Usage = func() {
		stderr.Printf("usage: %s [YOUTUBE_URL]", os.Args[0])
		stderr.Println()
		stderr.Printf("example:\n%s https://youtube.com/watch?v=oRdxUFDoQe0", os.Args[0])
		stderr.Println()

		stderr.Print("flags:")
		flag.PrintDefaults()
	}

	var skipVideo bool
	var skipAudio bool
	flag.BoolVar(&skipVideo, "nv", false, "skip downloading video")
	flag.BoolVar(&skipAudio, "na", false, "skip downloading audio")
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	url, err := urlpkg.Parse(flag.Arg(0))
	if err != nil {
		log.Panicf("invalid url: %v", err)
	}

	videoID := url.Query().Get("v")
	if videoID == "" {
		log.Panicf("url doesn't contain video parameter (?v=): %s", flag.Arg(0))
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
	fmt.Println()

	pipedSvc := piped.NewService(stderr, top[0].URL, http.DefaultClient)

	video, err := pipedSvc.FindStreams(ctx, videoID)
	if err != nil {
		stderr.Panic(err)
	}

	fmt.Println(video.Title)
	fmt.Println()
	time.Sleep(time.Second)

	lines := bufioctx.NewScanner(os.Stdin)

	var videoOption int
	if !skipVideo {
		fmt.Println("0) No video")
		for i, s := range video.VideoStreams {
			fps := "?"
			if s.FPS != 0 {
				fps = strconv.Itoa(s.FPS)
			}
			fmt.Printf("%d) %s - %s fps - %.1f MB - %s\n", i+1, s.Quality, fps, float64(s.ContentLength)/1024/1024, s.Codec)
		}
		fmt.Print("Choose video option (0 to skip): ")
		lines.Scan(ctx)
		videoOption, err = strconv.Atoi(lines.Text())
		if err != nil {
			stderr.Panic("invalid index")
		}
		if videoOption < 0 || videoOption > len(video.VideoStreams) {
			stderr.Panic("invalid index")
		}
		fmt.Println()
	}

	var audioOption int
	if !skipAudio {
		fmt.Println("0) No audio")
		for i, s := range video.AudioStreams {
			fmt.Printf("%d) %s - %s - %s\n", i+1, s.Quality, s.Codec, s.MIMEType)
		}
		fmt.Print("Choose audio option (0 to skip): ")
		lines.Scan(ctx)
		audioOption, err = strconv.Atoi(lines.Text())
		if err != nil {
			stderr.Panic("invalid index")
		}
		if audioOption < 0 || audioOption > len(video.AudioStreams) {
			stderr.Panic("invalid index")
		}
		fmt.Println()
	}

	var videoStream *piped.VideoStream
	if videoOption != 0 {
		videoStream = &video.VideoStreams[videoOption-1]
	}

	var audioStream *piped.AudioStream
	if audioOption != 0 {
		audioStream = &video.AudioStreams[audioOption-1]
	}

	mediaSvc := media.NewFFmpegService()
	downloadSvc := download.NewService(http.DefaultClient)
	err = downloadSvc.DownloadFromPiped(ctx, download.DownloadFromPipedParams{
		Title:        video.Title,
		AudioStream:  audioStream,
		VideoStream:  videoStream,
		MediaService: mediaSvc,
	})
	if err != nil {
		stderr.Panic(err)
	}

	stderr.Print("finished")
}
