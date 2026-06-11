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
	replyCountRegex    = regexp.MustCompile(`"?(?:reply_count)"?:\s?(\d+)`)
	favoriteCountRegex = regexp.MustCompile(`"?(?:favorite_count)"?:\s?(\d+)`)
	retweetCountRegex  = regexp.MustCompile(`"?(?:retweet_count)"?:\s?(\d+)`)
	bookmarkCountRegex = regexp.MustCompile(`"?(?:bookmark_count)"?:\s?(\d+)`)
	quoteCountRegex    = regexp.MustCompile(`"?(?:quote_count)"?:\s?(\d+)`)
	// viewCountRegex = regexp.MustCompile(`"?(?:view_count|views_count)"?:\s?"?(\d+)"?`)
	// textViewRegex  = regexp.MustCompile(`([\d,.]+)\s*Views`)
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
		// Check for LD+JSON
		d.Find("script[type='application/ld+json']").Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Text(), "SocialMediaPosting") {
				found = true
			}
		})
		// Or INITIAL_STATE
		if !found {
			d.Find("script").Each(func(i int, s *goquery.Selection) {
				if strings.Contains(s.Text(), "__INITIAL_STATE__") {
					found = true
				}
			})
		}
		return found
	}

	if err != nil || !hasData(doc) {
		fmt.Printf("\nmethod1 fail %s, trying method2\n", link)
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

	htmlContent, _ := doc.Html()

	// If JSON-LD failed, try to get basic info from metadata
	if !jsonFound {
		ldData.URL = link
		ldData.ArticleBody = doc.Find("meta[property='og:description']").AttrOr("content", "")
		ldData.Author.Name = doc.Find("meta[property='og:title']").AttrOr("content", "")
		// AlternateName from Title usually is "Name (@handle) on X"
		title := doc.Find("title").Text()
		if matches := regexp.MustCompile(`\((@\w+)\)`).FindStringSubmatch(title); len(matches) > 1 {
			ldData.Author.AlternateName = matches[1]
		}
	}

	extractCount := func(re *regexp.Regexp, content string) uint32 {
		match := re.FindStringSubmatch(content)
		if len(match) > 1 {
			val := strings.ReplaceAll(match[1], ",", "")
			val = strings.ReplaceAll(val, ".", "")
			var count uint32
			fmt.Sscanf(val, "%d", &count)
			return count
		}
		return 0
	}

	// extractCount64 := func(re *regexp.Regexp, content string) uint64 {
	// 	match := re.FindStringSubmatch(content)
	// 	if len(match) > 1 {
	// 		val := strings.ReplaceAll(match[1], ",", "")
	// 		val = strings.ReplaceAll(val, ".", "")
	// 		var count uint64
	// 		fmt.Sscanf(val, "%d", &count)
	// 		return count
	// 	}
	// 	return 0
	// }

	replies := extractCount(replyCountRegex, htmlContent)
	likes := extractCount(favoriteCountRegex, htmlContent)
	retweets := extractCount(retweetCountRegex, htmlContent)
	bookmarks := extractCount(bookmarkCountRegex, htmlContent)
	quotes := extractCount(quoteCountRegex, htmlContent)
	// views := extractCount64(viewCountRegex, htmlContent)

	// if views == 0 {
	// 	views = extractCount64(textViewRegex, htmlContent)
	// }

	_ = bookmarks

	dt, err := time.Parse(time.RFC3339, ldData.DatePublished)
	if err != nil {
		dt = time.Now()
	}

	username := strings.TrimPrefix(ldData.Author.AlternateName, "@")

	dataTwit := &models.SocialData{
		Year:   uint16(dt.Year()),
		Month:  uint8(dt.Month()),
		Day:    uint8(dt.Day()),
		Text:   ldData.ArticleBody,
		Hour:   uint8(dt.Hour()),
		Minute: uint8(dt.Minute()),
		Title:  ldData.Author.Name,
		URL:    link,
		Author: username,
		// Views:     views,
		Likes:     likes,
		Comments:  replies,
		Repost:    retweets,
		Quotes:    quotes,
		SumRepost: retweets + quotes,
		Created:   dt,
	}

	return dataTwit, nil
}
