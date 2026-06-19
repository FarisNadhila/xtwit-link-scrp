package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xtwit-link-scrp/config"
	"xtwit-link-scrp/models"
)

var userAgents = map[string]string{
	"iPhone":  "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
	"Desktop": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"iPad":    "Mozilla/5.0 (iPad; CPU OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
}

func match(pattern, s string) string {
	m := regexp.MustCompile(pattern).FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

func parseNum(s string) uint64 {
	s = strings.ToUpper(strings.ReplaceAll(s, ",", ""))
	multiplier := float64(1)
	if strings.HasSuffix(s, "K") {
		multiplier = 1000
		s = strings.TrimSuffix(s, "K")
	} else if strings.HasSuffix(s, "M") {
		multiplier = 1000000
		s = strings.TrimSuffix(s, "M")
	}
	val, _ := strconv.ParseFloat(s, 64)
	return uint64(val * multiplier)
}

func findScript(html, postID string) string {
	for _, m := range regexp.MustCompile(`(?s)<script[^>]*>(.*?)</script>`).FindAllStringSubmatch(html, -1) {
		if strings.Contains(m[1], postID) && (strings.Contains(m[1], "likers") || strings.Contains(m[1], "creation_time")) {
			return m[1]
		}
	}
	return ""
}

func ScrapeFacebook(link string) (*models.SocialData, error) {
	postID := match(`(?:reel|reels|posts|videos|permalink|photo\.php|story_fbid|fbid)[=/]([0-9]+)`, link)
	if postID == "" {
		return nil, fmt.Errorf("invalid facebook URL: %s", link)
	}

	// htmlIPhoneDoc, _ := config.RequestHeadlessWithUA(link, 10, userAgents["iPhone"])
	htmlDesktopDoc, _ := config.RequestHeadlessWithUA(link, 10, userAgents["Desktop"])
	htmlIPadDoc, _ := config.RequestHeadlessWithUA(link, 10, userAgents["iPad"])

	// htmlIPhone, _ := htmlIPhoneDoc.Html()
	htmlDesktop, _ := htmlDesktopDoc.Html()
	htmlIPad, _ := htmlIPadDoc.Html()

	scriptDesktop := findScript(htmlDesktop, postID)
	scriptIPad := findScript(htmlIPad, postID)

	ogTitle := match(`<meta property="og:title" content="([^"]+)"`, htmlDesktop)
	ogURL := match(`<meta property="og:url" content="([^"]+)"`, htmlDesktop)

	views := parseNum(match(`([\d\.,KMkm]+)\s*views`, ogTitle))
	likes := parseNum(match(`"likers":\{"count":(\d+)\}`, scriptDesktop))
	comments := parseNum(match(`"total_comment_count":(\d+)`, scriptDesktop))
	shares := parseNum(match(`"share_count_reduced":"(\d+)"`, scriptDesktop))
	createdTime := parseNum(match(`"creation_time":(\d+)`, scriptDesktop))
	message := match(`"message":\{"text":"(.*?)"\}`, scriptDesktop)

	// Fallback to iPad or iPhone data if desktop is empty
	if likes == 0 {
		likes = parseNum(match(`"likers":\{"count":(\d+)\}`, scriptIPad))
	}
	if likes == 0 {
		likes = parseNum(match(`([\d\.,KMkm]+)\s*reactions`, ogTitle))
	}
	if comments == 0 {
		comments = parseNum(match(`"total_comment_count":(\d+)`, scriptIPad))
	}

	author := ""
	if parts := strings.Split(strings.TrimPrefix(ogURL, "https://www.facebook.com/"), "/"); len(parts) > 0 {
		author = parts[0]
	}

	t := time.Unix(int64(createdTime), 0)

	data := &models.SocialData{
		URL:      link,
		Author:   author,
		Title:    author,
		Likes:    uint32(likes),
		Comments: uint32(comments),
		Share:    uint32(shares),
		Views:    views,
		Created:  t,
		Year:     uint16(t.Year()),
		Month:    uint8(t.Month()),
		Day:      uint8(t.Day()),
		Hour:     uint8(t.Hour()),
		Minute:   uint8(t.Minute()),
		Text:     strings.ReplaceAll(strings.ReplaceAll(message, `\n`, "\n"), `\"`, `"`),
	}

	return data, nil
}
