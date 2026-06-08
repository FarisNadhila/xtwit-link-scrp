package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SocialData struct {
	Year        uint16
	Month       uint8
	Day         uint8
	Text        string
	Hour        uint8
	Minute      uint8
	Title       string
	URL         string
	Author      string
	Views       uint64
	Likes       uint32
	Repost      uint32
	Quotes      uint32
	TotalRepost uint32
	Comments    uint32
	Share       uint32
	Following   uint16
	Followers   uint32
	Created     time.Time
}

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

func main() {
	var link string = "https://www.threads.com/@dimas.budiartanto/post/DZNKXgjEt40"

	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Timeout: 20 * time.Second,
		Jar:     jar,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Fatal("error")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	// req.Header.Set("Cookie", "csrftoken=gqUf_nV5kEp01aSL1lklDf")

	// req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	// req.Header.Set("Upgrade-Insecure-Requests", "1")
	// req.Header.Set("Sec-Fetch-Dest", "document")
	// req.Header.Set("Sec-Fetch-Site", "none")
	// req.Header.Set("Sec-Fetch-User", "?1")
	// req.Header.Set("Priority", "u=0, i")

	// +"ig_did=D5CB5F6D-ADD5-4374-99DD-1E1995EF2F96; "
	// +"mid=ahV4ggAEAAG5Ewy8AR4KmExjyRLD; "
	// +"dpr=1.25; ps_l=1; ps_n=1"

	// req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error: %v", resp.StatusCode)
	}

	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return
	// }

	// os.WriteFile("tests/threads/response.html", bodyBytes, 0644)

	// if bytes.Contains(bodyBytes, []byte("direct_reply_count")) {
	// 	fmt.Println("FOUND")
	// } else {
	// 	fmt.Println("NOT FOUND")
	// }

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("error")
	}

	var jsonText string

	doc.Find(`script[type="application/json"]`).Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		if strings.Contains(text, "direct_reply_count") {
			jsonText = text
		}
	})

	var result map[string]any

	err = json.Unmarshal([]byte(jsonText), &result)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	username := GetString(result, "username")
	fullName := GetString(result, "full_name")

	likes := GetInt(result, "like_count")
	replies := GetInt(result, "direct_reply_count")
	reposts := GetInt(result, "repost_count")
	quotes := GetInt(result, "quote_count")
	reshares := GetInt(result, "reshare_count")

	takenAt := GetInt(result, "taken_at")

	created := time.Unix(int64(takenAt), 0)

	social := SocialData{
		Title:       fullName,
		Author:      username,
		URL:         link,
		Likes:       uint32(likes),
		Comments:    uint32(replies),
		Share:       uint32(reshares),
		Repost:      uint32(reposts),
		Quotes:      uint32(quotes),
		TotalRepost: uint32(reposts + quotes),
		Created:     created,
		Year:        uint16(created.Year()),
		Month:       uint8(created.Month()),
		Day:         uint8(created.Day()),
		Hour:        uint8(created.Hour()),
		Minute:      uint8(created.Minute()),
	}

	fmt.Println(social)
}
