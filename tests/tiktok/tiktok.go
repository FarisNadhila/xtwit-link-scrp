package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"xtwit-link-scrp/models"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var link string = "https://www.tiktok.com/@caliakcaliak2/video/7627329973765917959"

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error: %v %v", resp.StatusCode, err)
	}

	// write html
	// bodyBytes, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	fmt.Printf("error: %v", err)
	// }

	// os.WriteFile("tests/tiktok/response.html", bodyBytes, 0644)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	s := doc.Find(`script[type="application/json"]#api-data`)
	if s.Length() == 0 {
		fmt.Printf("error: %v", err)
	}

	// write json
	jsonText := strings.TrimSpace(s.Text())
	os.WriteFile("tests/tiktok/data.json", []byte(jsonText), 0644)

	var result map[string]any

	err = json.Unmarshal([]byte(jsonText), &result)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	videoDet := result["videoDetail"].(map[string]any)
	itemInfo := videoDet["itemInfo"].(map[string]any)
	itemStruct := itemInfo["itemStruct"].(map[string]any)
	author := itemStruct["author"].(map[string]any)
	stats := itemStruct["stats"].(map[string]any)
	createTimeStr := itemStruct["createTime"].(string)

	ts, err := strconv.ParseInt(createTimeStr, 10, 64)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	t := time.Unix(ts, 0)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tID := t.In(loc)

	data := &models.SocialData{
		Author:    author["uniqueId"].(string),
		Title:     author["nickname"].(string),
		Followers: uint32(author["followerCount"].(float64)),
		Following: uint16(author["followingCount"].(float64)),
		Likes:     uint32(stats["diggCount"].(float64)),
		Comments:  uint32(stats["commentCount"].(float64)),
		Share:     uint32(stats["shareCount"].(float64)),
		Views:     uint64(stats["playCount"].(float64)),
		Text:      itemStruct["desc"].(string),
		URL:       link,
		Created:   tID,
	}

	data.Year = uint16(data.Created.Year())
	data.Month = uint8(data.Created.Month())
	data.Day = uint8(data.Created.Day())
	data.Hour = uint8(data.Created.Hour())
	data.Minute = uint8(data.Created.Minute())

	fmt.Println(data)

}
