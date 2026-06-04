package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
	"xtwit-link-scrp/models"

	"github.com/PuerkitoBio/goquery"
)

var (
	initialStateRegex = regexp.MustCompile(`window\.__INITIAL_STATE__\s*=\s*({.*?});`)
)

type TwitterInitialState struct {
	Entities struct {
		Tweets struct {
			Entities map[string]interface{} `json:"entities"`
		} `json:"tweets"`
		Users struct {
			Entities map[string]interface{} `json:"entities"`
		} `json:"users"`
	} `json:"entities"`
}

func ScrapeTwitter(link string) (*models.TweetData, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	time.Sleep(1 * time.Second)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, err
	}

	var jsonStr string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "window.__INITIAL_STATE__") {
			match := initialStateRegex.FindStringSubmatch(text)
			if len(match) > 1 {
				jsonStr = match[1]
			}
		}
	})

	if jsonStr == "" {
		match := initialStateRegex.FindStringSubmatch(string(bodyBytes))
		if len(match) > 1 {
			jsonStr = match[1]
		}
	}

	if jsonStr == "" {
		return nil, fmt.Errorf("could not find INITIAL_STATE")
	}

	var data TwitterInitialState
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, err
	}

	parts := strings.Split(link, "/")
	tweetID := strings.Split(parts[len(parts)-1], "?")[0]

	tweetEntities := data.Entities.Tweets.Entities
	var tweetData map[string]interface{}

	if val, ok := tweetEntities[tweetID]; ok {
		tweetData = val.(map[string]interface{})
	} else {
		for k, v := range tweetEntities {
			if k == tweetID {
				tweetData = v.(map[string]interface{})
				break
			}
		}
	}

	if tweetData == nil {
		return nil, fmt.Errorf("tweet data not found")
	}

	userID := fmt.Sprintf("%v", tweetData["user"])
	userEntities := data.Entities.Users.Entities
	userData, _ := userEntities[userID].(map[string]interface{})

	text, _ := tweetData["full_text"].(string)
	if text == "" {
		text, _ = tweetData["text"].(string)
	}

	username, _ := userData["screen_name"].(string)
	accountName, _ := userData["name"].(string)
	createdAt, _ := tweetData["created_at"].(string)
	likes, _ := tweetData["favorite_count"].(float64)

	dt, err := time.Parse("2006-01-02T15:04:05.000Z", createdAt)
	if err != nil {
		dt = time.Now()
	}

	return &models.TweetData{
		Year:    dt.Year(),
		Month:   int(dt.Month()),
		Day:     dt.Day(),
		Text:    text,
		Hour:    dt.Hour(),
		Minute:  dt.Minute(),
		Title:   accountName,
		URL:     link,
		Author:  username,
		Views:   0,
		Likes:   int(likes),
		Created: dt,
	}, nil
}
