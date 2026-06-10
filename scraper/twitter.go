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
	replyCountRegex    = regexp.MustCompile(`reply_count:(\d+)`)
	favoriteCountRegex = regexp.MustCompile(`favorite_count:(\d+)`)
	retweetCountRegex  = regexp.MustCompile(`retweet_count:(\d+)`)
	bookmarkCountRegex = regexp.MustCompile(`bookmark_count:(\d+)`)
	quoteCountRegex    = regexp.MustCompile(`quote_count:(\d+)`)
	viewCountRegex     = regexp.MustCompile(`count:"(\d+)"`)
)

type TwitterLDJSON struct {
	ArticleBody string `json:"articleBody"`
	Author      struct {
		Name          string `json:"name"`
		AlternateName string `json:"alternateName"`
		URL           string `json:"url"`
	} `json:"author"`
	DatePublished string `json:"datePublished"`
	Headline      string `json:"headline"`
	Identifier    string `json:"identifier"`
	Image         string `json:"image"`
	URL           string `json:"url"`
}

func ScrapeTwitter(link string) (*models.SocialData, error) {
	doc, err := config.RequestStatic(link)

	hasData := func(d *goquery.Document) bool {
		found := false
		d.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "SocialMediaPosting") {
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

	var ldData TwitterLDJSON
	var jsonFound bool
	doc.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
		var temp map[string]interface{}
		if err := json.Unmarshal([]byte(s.Text()), &temp); err == nil {
			if temp["@type"] == "SocialMediaPosting" {
				if b, err := json.Marshal(temp); err == nil {
					json.Unmarshal(b, &ldData)
					jsonFound = true
				}
			}
		}
	})

	if !jsonFound {
		return nil, fmt.Errorf("could not find SocialMediaPosting JSON-LD")
	}

	htmlContent, _ := doc.Html()

	extractCount := func(re *regexp.Regexp, content string) uint32 {
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			var count uint32
			fmt.Sscanf(match[1], "%d", &count)
			return count
		}
		return 0
	}

	extractCount64 := func(re *regexp.Regexp, content string) uint64 {
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			var count uint64
			fmt.Sscanf(match[1], "%d", &count)
			return count
		}
		return 0
	}

	replies := extractCount(replyCountRegex, htmlContent)
	likes := extractCount(favoriteCountRegex, htmlContent)
	retweets := extractCount(retweetCountRegex, htmlContent)
	bookmarks := extractCount(bookmarkCountRegex, htmlContent)
	quotes := extractCount(quoteCountRegex, htmlContent)
	views := extractCount64(viewCountRegex, htmlContent)

	_ = bookmarks

	dt, err := time.Parse(time.RFC3339, ldData.DatePublished)
	if err != nil {
		dt = time.Now()
	}

	username := strings.TrimPrefix(ldData.Author.AlternateName, "@")

	dataTwit := &models.SocialData{
		Year:      uint16(dt.Year()),
		Month:     uint8(dt.Month()),
		Day:       uint8(dt.Day()),
		Text:      ldData.ArticleBody,
		Hour:      uint8(dt.Hour()),
		Minute:    uint8(dt.Minute()),
		Title:     ldData.Author.Name,
		URL:       link,
		Author:    username,
		Views:     views,
		Likes:     likes,
		Comments:  replies,
		Repost:    retweets,
		Quotes:    quotes,
		SumRepost: retweets + quotes,
		Created:   dt,
	}

	return dataTwit, nil
}
