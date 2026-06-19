package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

type SocialData struct {
	URL      string
	Author   string
	Title    string
	Likes    uint32
	Comments uint32
	Share    uint32
	Views    uint64
	Created  time.Time
	Text     string
}

var userAgents = map[string]string{
	"iPhone":  "Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
	"Desktop": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
	"iPad":    "Mozilla/5.0 (iPad; CPU OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
}

func RequestHeadless(link string, ua string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent(ua),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(link),
		chromedp.Sleep(10*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	return htmlContent, err
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

func ScrapeFacebook(link string) (*SocialData, error) {
	postID := match(`(?:reel|reels|posts|videos|permalink|photo\.php|story_fbid|fbid)[=/]([0-9]+)`, link)
	if postID == "" {
		return nil, fmt.Errorf("invalid facebook URL: %s", link)
	}

	htmlIPhone, _ := RequestHeadless(link, userAgents["iPhone"])
	htmlDesktop, _ := RequestHeadless(link, userAgents["Desktop"])
	htmlIPad, _ := RequestHeadless(link, userAgents["iPad"])

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

	return &SocialData{
		URL:      link,
		Author:   author,
		Title:    author,
		Likes:    uint32(likes),
		Comments: uint32(comments),
		Share:    uint32(shares),
		Views:    views,
		Created:  t,
		Text:     strings.ReplaceAll(strings.ReplaceAll(message, `\n`, "\n"), `\"`, `"`),
	}, nil
}

func main() {
	link1 := "https://www.facebook.com/reel/788180107420485"
	data, err := ScrapeFacebook(link1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("\n--- FINAL PARSED DATA ---")
	fmt.Printf("URL:         %s\n", data.URL)
	fmt.Printf("Account ID:  %s\n", data.Author)
	fmt.Printf("Account Name:%s\n", data.Title)
	fmt.Printf("Likes:       %d\n", data.Likes)
	fmt.Printf("Comments:    %d\n", data.Comments)
	fmt.Printf("Shares:      %d\n", data.Share)
	fmt.Printf("Views:       %d\n", data.Views)
	fmt.Printf("Created:     %v\n", data.Created)
	fmt.Printf("Text:        %s\n", strings.ReplaceAll(data.Text, "\n", " "))
}
