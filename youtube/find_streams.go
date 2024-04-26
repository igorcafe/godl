package youtube

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

// curl 'https://api.piped.projectsegfau.lt/streams/YAmLMohrus4' \
//   -H 'sec-ch-ua: "Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"' \
//   -H 'Referer: https://piped.projectsegfau.lt/' \
//   -H 'sec-ch-ua-mobile: ?0' \
//   -H 'User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36' \
//   -H 'sec-ch-ua-platform: "Linux"'

type VideoStream struct {
	URL           string
	Format        string
	ContentLength int
	Quality       string
	FPS           int
	VideoOnly     bool
}

type AudioStream struct {
	URL           string
	Format        string
	ContentLength int
	Quality       string
}

type StreamsResponse struct {
	VideoStreams []VideoStream
	AudioStreams []AudioStream
}

func (s service) FindStreams(ctx context.Context, videoID string) (StreamsResponse, error) {
	var streams StreamsResponse
	//	  -H 'authority: api.piped.projectsegfau.lt' \
	//	  -H 'origin: https://piped.projectsegfau.lt' \
	//	  -H 'referer: https://piped.projectsegfau.lt/' \
	//	  -H 'user-agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36'

	url := s.baseURL + "/streams/" + videoID
	log.Print(url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return streams, err
	}

	req.Header.Set("Origin", s.baseURL)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")

	resp, err := s.http.Do(req)
	if err != nil {
		return streams, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return streams, err
	}

	err = json.Unmarshal(b, &streams)
	if err != nil {
		log.Print(resp.Status)
		log.Print(string(b))
		return streams, err
	}

	return streams, nil
}
