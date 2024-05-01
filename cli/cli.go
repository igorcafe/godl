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
		stderr.Panic("no videoID")
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

	pipedSvc := piped.NewService(top[0].URL, http.DefaultClient)

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

	dlSvc := download.NewService(http.DefaultClient)
	errs := make(chan error)
	downloadCount := 0

	if videoOption != 0 {
		downloadCount++
		go func() {
			videoPath := "video.mp4"
			videoStream := video.VideoStreams[videoOption-1]
			now := time.Now()
			stderr.Print("video: download started")
			err := dlSvc.DownloadStream(ctx, videoStream.URL, videoPath, func(elapsed, total int64) {
				stderr.Printf("video: %d KB / %d KB\n", elapsed/1024, total/1024)
			})
			if err == nil {
				stderr.Printf("video: finished downloading in %.1f seconds", time.Since(now).Seconds())
			}
			errs <- err
		}()
	}

	if audioOption != 0 {
		downloadCount++
		go func() {
			audioPath := video.Title + ".mp3"
			audioStream := video.AudioStreams[audioOption-1]
			now := time.Now()
			stderr.Print("audio: download started")
			err := dlSvc.DownloadStream(ctx, audioStream.URL, audioPath, func(elapsed, total int64) {
				stderr.Printf("audio: %d KB / %d KB\n", elapsed/1024, total/1024)
			})
			if err == nil {
				fmt.Printf("audio: finished downloading in %.1f seconds\n", time.Since(now).Seconds())
			}
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
