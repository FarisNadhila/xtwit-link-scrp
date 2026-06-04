package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	link := "https://www.instagram.com/p/DVr-al0FkWh/"

	client := &http.Client{Timeout: 20 * time.Second}
	req, _ := http.NewRequest("GET", link, nil)
	// req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")
	req.Header.Set("Cookie", "csrftoken=NmUEcXmldw8DOADSZcxyrw; datr=6HYVaueHz82EXF_LQ4A8jppv; ig_did=60563026-08A3-41FC-A38D-2FF3E0E1503C; mid=ahV26AAEAAHHWwuPFMZuqrA0NNjx; ig_nrcb=1; wd=780x816; dpr=1.25; ps_l=1; ps_n=1")
	req.Header.Set("Accept", "*/*")
	// req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Referer", link)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Request failed:", err)
		return
	}
	defer resp.Body.Close()

	var reader io.Reader
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Println("Gzip error:", err)
			return
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = resp.Body
	}

	bodyBytes, _ := io.ReadAll(reader)
	os.WriteFile("response.html", bodyBytes, 0644)

	fmt.Println("Status:", resp.StatusCode)

	// bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	doc, _ := goquery.NewDocumentFromReader(strings.NewReader(body))

	found := false
	doc.Find("time").Each(func(i int, s *goquery.Selection) {
		dt, exists := s.Attr("datetime")
		if exists && dt != "" {
			fmt.Println("<time datetime>:", dt)
			if t, err := time.Parse(time.RFC3339, dt); err == nil {
				fmt.Println("   Parsed:", t.UTC())
			}
			found = true
		}
	})

	if !found {
		fmt.Println("No <time datetime> found")
		fmt.Println("Resp snippet:", body[:500])
	}
}
