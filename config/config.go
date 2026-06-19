package config

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
)

var (
	sharedClient *http.Client
	userAgents   = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (iPad; CPU OS 17_4_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.4.1 Mobile/15E148 Safari/604.1",
		// "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
		// "Mozilla/5.0 (compatible; bingbot/2.0; +http://www.bing.com/bingbot.htm)",
		// "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Twitterbot/1.0",
		// "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)",
	}
)

func init() {
	jar, _ := cookiejar.New(nil)
	sharedClient = &http.Client{
		Timeout: 30 * time.Second,
		Jar:     jar,
	}
}

func RequestStatic(link string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	ua := userAgents[rand.Intn(len(userAgents))]
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)

	resp, err := sharedClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %v", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func RequestHeadless(link string, timeout uint8) (*goquery.Document, error) {
	ua := userAgents[rand.Intn(len(userAgents))]
	return RequestHeadlessWithUA(link, timeout, ua)
}

func RequestHeadlessWithUA(link string, timeout uint8, ua string) (*goquery.Document, error) {
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

	duration := time.Duration(timeout)

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(link),
		chromedp.Sleep(duration*time.Second),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func RequestLink(link string) (*goquery.Document, error) {
	doc, err := RequestStatic(link)
	if err == nil {
		return doc, nil
	}

	fmt.Printf("\nmethod1 fail %s, trying method2\n", link)
	return RequestHeadless(link, 10)
}
