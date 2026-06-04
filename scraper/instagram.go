package scraper

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"xtwit-link-scrp/models"

	"github.com/PuerkitoBio/goquery"
)

var (
	igLikesRegex    = regexp.MustCompile(`(?i)([\d,.]+)\s*like`)
	igCommentsRegex = regexp.MustCompile(`(?i)([\d,.]+)\s*comment`)
)

func ScrapeInstagram(link string) (*models.InstagramData, error) {
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1")

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

	data := &models.InstagramData{
		URL:     link,
		Created: time.Now(),
	}

	timeFromTag := false

	doc.Find("time").Each(func(i int, s *goquery.Selection) {
		dt, exists := s.Attr("datetime")
		if exists && dt != "" {
			if t, err := time.Parse(time.RFC3339, dt); err == nil {
				data.Created = t
				timeFromTag = true
			}
		}
	})

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		switch property {
		case "og:description":
			if match := igLikesRegex.FindStringSubmatch(content); len(match) > 1 {
				data.Likes = parseIGNumber(match[1])
			}
			if match := igCommentsRegex.FindStringSubmatch(content); len(match) > 1 {
				data.Comments = parseIGNumber(match[1])
			}

			parts := strings.SplitN(content, " - ", 2)
			if len(parts) > 1 {
				remain := parts[1]
				onParts := strings.SplitN(remain, " on ", 2)
				if len(onParts) > 1 {
					data.Author = onParts[0]

					dateCaptionParts := strings.SplitN(onParts[1], ": \"", 2)
					if len(dateCaptionParts) > 1 {
						dateStr := dateCaptionParts[0]
						caption := strings.TrimSuffix(dateCaptionParts[1], "\"")
						data.Text = caption

						formats := []string{"January 2, 2006", "Jan 2, 2006"}
						for _, fmtStr := range formats {
							if t, err := time.Parse(fmtStr, dateStr); err == nil {
								data.Created = t
								break
							}
						}
					}
				}
			}

			if data.Text == "" {
				data.Text = content
			}

		case "og:title":
			parts := strings.Split(content, " on Instagram")
			if len(parts) > 0 {
				data.Title = strings.Trim(parts[0], "\" ")
			}
		}
	})

	if data.Title == "" && data.Author != "" {
		data.Title = data.Author
	}
	if data.Author == "" && data.Title != "" {
		data.Author = data.Title
	}

	accountID := ""
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		patterns := []string{
			`"user_id":"(\d+)"`,
			`"id":"(\d+)"`,
			`"owner":\{"id":"(\d+)"`,
			`"profilePage_(\d+)"`,
		}
		for _, p := range patterns {
			re := regexp.MustCompile(p)
			match := re.FindStringSubmatch(text)
			if len(match) > 1 {
				accountID = match[1]
				break
			}
		}

		if !timeFromTag {
			tsRe := regexp.MustCompile(`"taken_at_timestamp":\s*(\d+)`)
			if tsMatch := tsRe.FindStringSubmatch(text); len(tsMatch) > 1 {
				if ts, err := strconv.ParseInt(tsMatch[1], 10, 64); err == nil {
					data.Created = time.Unix(ts, 0).UTC()
					timeFromTag = true
				}
			}
		}
		if !timeFromTag {
			dateRe := regexp.MustCompile(`"(uploadDate|datePublished)":"([^"]+)"`)
			if dateMatch := dateRe.FindStringSubmatch(text); len(dateMatch) > 2 {
				if t, err := time.Parse(time.RFC3339, dateMatch[2]); err == nil {
					data.Created = t
					timeFromTag = true
				}
			}
		}
	})

	if accountID != "" {
		data.Author = accountID
	} else if data.Author == "" {
		data.Author = data.Title
	}

	data.Year = data.Created.Year()
	data.Month = int(data.Created.Month())
	data.Day = data.Created.Day()
	data.Hour = data.Created.Hour()
	data.Minute = data.Created.Minute()

	return data, nil
}

func parseIGNumber(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	val, _ := strconv.Atoi(s)
	return val
}
