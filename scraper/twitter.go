package scraper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"xtwit-link-scrp/config"
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

func ScrapeTwitter(link string) (*models.SocialData, error) {
	doc, err := config.RequestStatic(link)

	hasData := func(d *goquery.Document) bool {
		found := false
		d.Find("script").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "window.__INITIAL_STATE__") {
				found = true
			}
		})
		return found
	}

	if err != nil || !hasData(doc) {
		fmt.Printf("method1 fail %s, trying method2\n", link)
		doc, err = config.RequestHeadless(link)
		if err != nil {
			return nil, err
		}
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

	dataTwit := &models.SocialData{
		Year:    uint16(dt.Year()),
		Month:   uint8(dt.Month()),
		Day:     uint8(dt.Day()),
		Text:    text,
		Hour:    uint8(dt.Hour()),
		Minute:  uint8(dt.Minute()),
		Title:   accountName,
		URL:     link,
		Author:  username,
		Views:   0,
		Likes:   uint32(likes),
		Created: dt,
	}

	return dataTwit, nil
}
