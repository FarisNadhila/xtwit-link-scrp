package scraper

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"xtwit-link-scrp/config"
	"xtwit-link-scrp/models"
)

func ScrapeTiktok(link string) (*models.SocialData, error) {
	doc, err := config.RequestStatic(link)
	if err != nil || doc.Find(`script[type="application/json"]#api-data`).Length() == 0 {
		fmt.Printf("\nmethod1 fail %s, trying method2\n", link)
		doc, err = config.RequestHeadless(link, 5)
		if err != nil {
			return nil, err
		}
	}

	s := doc.Find(`script[type="application/json"]#api-data`)
	if s.Length() == 0 {
		return nil, fmt.Errorf("could not find tiktok api data")
	}

	jsonText := strings.TrimSpace(s.Text())

	var result map[string]any

	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, err
	}

	videoDet, _ := result["videoDetail"].(map[string]any)
	itemInfo, _ := videoDet["itemInfo"].(map[string]any)
	itemStruct, _ := itemInfo["itemStruct"].(map[string]any)
	author, _ := itemStruct["author"].(map[string]any)
	stats, _ := itemStruct["stats"].(map[string]any)
	createTimeStr, _ := itemStruct["createTime"].(string)

	ts, err := strconv.ParseInt(createTimeStr, 10, 64)
	if err != nil {
		return nil, err
	}

	t := time.Unix(ts, 0)
	loc, _ := time.LoadLocation("Asia/Jakarta")
	tID := t.In(loc)

	dataTik := &models.SocialData{
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
		Year:      uint16(tID.Year()),
		Month:     uint8(tID.Month()),
		Day:       uint8(tID.Day()),
		Hour:      uint8(tID.Hour()),
		Minute:    uint8(tID.Minute()),
	}

	return dataTik, nil
}
