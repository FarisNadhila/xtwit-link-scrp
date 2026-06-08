package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type SocialData struct {
	Year      uint16
	Month     uint8
	Day       uint8
	Text      string
	Hour      uint8
	Minute    uint8
	Title     string
	URL       string
	Author    string
	Views     uint64
	Likes     uint32
	Comments  uint32
	Share     uint32
	Following uint16
	Followers uint32
	Created   time.Time
}

func main() {
	var link string = "https://www.instagram.com/p/DVr-al0FkWh/"

	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		log.Fatal("error")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 14_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0.3 Mobile/15E148 Safari/604.1")

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
	// 	log.Fatal("error")
	// }

	// os.WriteFile("tests/instagram/response.html", bodyBytes, 0644)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("error")
	}

	s := doc.Find(`meta[name="description"]`)       // likes, comments, date
	sTitle := doc.Find(`meta[property="og:title"]`) // caption, username, account name
	titlesm, _ := sTitle.Attr("content")
	content, _ := s.Attr("content")

	// fmt.Printf("titlelism: %v\n", titlesm)
	reLikes := regexp.MustCompile(`^([\d,]+) likes`)
	likes := reLikes.FindStringSubmatch(content)

	reComments := regexp.MustCompile(`([\d,]+) comments`)
	comments := reComments.FindStringSubmatch(content)

	reUsername := regexp.MustCompile(`comments - (\S+) on`)
	username := reUsername.FindStringSubmatch(content)

	reDate := regexp.MustCompile(`on ([A-Za-z]+ \d+, \d{4}):`)
	created := reDate.FindStringSubmatch(content)

	// reCaption := regexp.MustCompile(`(?s):\s*"(.+)"\s*\.\s*$`)
	// caption := reCaption.FindStringSubmatch(content)

	reCaption := regexp.MustCompile(`^.*? on Instagram:\s*`)
	caption := reCaption.ReplaceAllString(titlesm, "")

	reAccountName := regexp.MustCompile(`^\s*(.*?)\s+on Instagram:`)
	AccountName := reAccountName.FindStringSubmatch(titlesm)
	// AccountName, _, _ := strings.Cut(reAccountName[0], " on Instagram:")

	t, err := time.Parse("January 2, 2006", created[1])
	if err != nil {
		log.Fatal("error")
	}

	parseNum := func(s string) uint32 {
		s = strings.ReplaceAll(s, ",", "")
		n, _ := strconv.ParseUint(s, 10, 32)
		return uint32(n)
	}

	data := SocialData{
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

	fmt.Println(data)

	// fmt.Printf("Account name: %v\n", AccountName[1])
	// fmt.Printf("likes: %v\n", parseNum(likes[1]))
	// fmt.Printf("comment: %v\n", parseNum(comments[1]))
	// fmt.Printf("username: %v\n", username[1])
	// fmt.Printf("caption: %v\n", caption)
	// fmt.Printf("created: %v\n", t)
	// fmt.Printf("day: %v\n", t.Day())
	// fmt.Printf("month: %v\n", int8(t.Month()))
	// fmt.Printf("year: %v\n", t.Year())
	// fmt.Printf("link: %v", link)

	// if err != nil {
	// 	fmt.Printf("Error: %v", err)
	// }

}
