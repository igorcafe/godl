package piped

import (
	"context"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func (s instanceService) List(ctx context.Context) ([]Instance, error) {
	url := "https://github.com/TeamPiped/Piped/wiki/Instances"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	instances := []Instance{}
	doc.Find("tr td:nth-child(2) a").Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr("href", "")
		if url == "" {
			return
		}
		instances = append(instances, Instance{
			URL: url,
		})
	})

	return instances, nil
}
