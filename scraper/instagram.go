package scraper

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"xtwit-link-scrp/config"
	"xtwit-link-scrp/models"
)

func ScrapeInstagram(link string) (*models.SocialData, error) {
	doc, err := config.RequestStatic(link)
	if err != nil || doc.Find(`meta[name="description"]`).Length() == 0 {
		log.Fatalf("\nmethod1 fail %s, trying method2\n", link)
		doc, err = config.RequestHeadless(link)
		if err != nil {
			return nil, err
		}
	}

	s := doc.Find(`meta[name="description"]`)       // likes, comments, date
	sTitle := doc.Find(`meta[property="og:title"]`) // caption, username, account name
	titlesm, _ := sTitle.Attr("content")
	content, _ := s.Attr("content")

	reLikes := regexp.MustCompile(`^([\d,]+) likes`)
	likes := reLikes.FindStringSubmatch(content)

	reComments := regexp.MustCompile(`([\d,]+) comments`)
	comments := reComments.FindStringSubmatch(content)

	reUsername := regexp.MustCompile(`comments - (\S+) on`)
	username := reUsername.FindStringSubmatch(content)

	reDate := regexp.MustCompile(`on ([A-Za-z]+ \d+, \d{4}):`)
	created := reDate.FindStringSubmatch(content)

	reCaption := regexp.MustCompile(`^.*? on Instagram:\s*`)
	caption := reCaption.ReplaceAllString(titlesm, "")

	reAccountName := regexp.MustCompile(`^\s*(.*?)\s+on Instagram:`)
	AccountName := reAccountName.FindStringSubmatch(titlesm)

	if len(created) < 2 {
		return nil, fmt.Errorf("failed to parse instagram date")
	}

	t, err := time.Parse("January 2, 2006", created[1])
	if err != nil {
		return nil, err
	}

	parseNum := func(s string) uint32 {
		s = strings.ReplaceAll(s, ",", "")
		n, _ := strconv.ParseUint(s, 10, 32)
		return uint32(n)
	}

	dataIG := &models.SocialData{
		Title:    AccountName[1],
		URL:      link,
		Author:   username[1],
		Likes:    parseNum(likes[1]),
		Comments: parseNum(comments[1]),
		Text:     caption,
		Created:  t,
		Year:     uint16(t.Year()),
		Month:    uint8(t.Month()),
		Day:      uint8(t.Day()),
	}

	return dataIG, nil
}
