package scraper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"xtwit-link-scrp/config"
	"xtwit-link-scrp/models"
)

func FindValue(data any, targetKey string) any {
	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			if key == targetKey {
				return value
			}
			if result := FindValue(value, targetKey); result != nil {
				return result
			}
		}
	case []any:
		for _, item := range v {

			if result := FindValue(item, targetKey); result != nil {
				return result
			}
		}
	}
	return nil
}

func GetString(data any, key string) string {
	if v := FindValue(data, key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetInt(data any, key string) int {
	if v := FindValue(data, key); v != nil {
		if n, ok := v.(float64); ok {
			return int(n)
		}
	}
	return 0
}

func ScrapeThreads(link string) (*models.SocialData, error) {
	doc, err := config.RequestStatic(link)

	hasJSON := func(d *goquery.Document) bool {
		found := false
		d.Find(`script[type="application/json"]`).Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "direct_reply_count") {
				found = true
			}
		})
		return found
	}

	if err != nil || !hasJSON(doc) {
		fmt.Printf("method1 fail %s, trying method2\n", link)
		doc, err = config.RequestHeadless(link)
		if err != nil {
			return nil, err
		}
	}

	var jsonText string

	doc.Find(`script[type="application/json"]`).Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		if strings.Contains(text, "direct_reply_count") {
			jsonText = text
		}
	})

	if jsonText == "" {
		return nil, fmt.Errorf("could not find JSON data in page")
	}

	var result map[string]any

	json.Unmarshal([]byte(jsonText), &result)

	username := GetString(result, "username")
	fullName := GetString(result, "full_name")
	caption := GetString(result, "text")

	likes := GetInt(result, "like_count")
	replies := GetInt(result, "direct_reply_count")
	reposts := GetInt(result, "repost_count")
	quotes := GetInt(result, "quote_count")
	reshares := GetInt(result, "reshare_count")

	takenAt := GetInt(result, "taken_at")
	created := time.Unix(int64(takenAt), 0)

	dataThreads := &models.SocialData{
		Title:     fullName,
		Author:    username,
		Text:      caption,
		URL:       link,
		Likes:     uint32(likes),
		Comments:  uint32(replies),
		Share:     uint32(reshares),
		Repost:    uint32(reposts),
		Quotes:    uint32(quotes),
		SumRepost: uint32(reposts + quotes),
		Created:   created,
		Year:      uint16(created.Year()),
		Month:     uint8(created.Month()),
		Day:       uint8(created.Day()),
		Hour:      uint8(created.Hour()),
		Minute:    uint8(created.Minute()),
	}

	return dataThreads, nil
}
