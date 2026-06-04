package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"xtwit-link-scrp/excel"
	"xtwit-link-scrp/models"
	"xtwit-link-scrp/scraper"

	"github.com/schollz/progressbar/v3"
)

const (
	InputFile    = "input.xlsx"
	TemplateFile = "template.xlsx"
	OutputFile   = "output.xlsx"
)

func main() {
	links, err := excel.ReadLinks(InputFile)
	if err != nil {
		log.Fatalf("Error reading links: %v", err)
	}

	results := runScraper(links)

	if len(results) == 0 {
		fmt.Println("No data scraped.")
		return
	}

	fmt.Printf("Writing %d results to %s...\n", len(results), OutputFile)
	if err := excel.WriteResults(TemplateFile, OutputFile, results); err != nil {
		log.Fatalf("Error writing results: %v", err)
	}

	fmt.Println("Success!")
}

func runScraper(links []string) []models.TweetData {
	bar := progressbar.Default(int64(len(links)), "Scraping")
	resultChan := make(chan models.TweetData, len(links))
	sem := make(chan struct{}, 50)
	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)
		go func(l string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			var result models.TweetData
			if strings.Contains(l, "twitter.com") || strings.Contains(l, "x.com") {
				var data *models.TweetData
				var err error
				for i := 0; i < 3; i++ {
					data, err = scraper.ScrapeTwitter(l)
					if err == nil && data != nil {
						break
					}
				}
				if data != nil {
					result = *data
				} else {
					result = models.TweetData{URL: l}
				}
			} else if strings.Contains(l, "instagram.com") {
				var data *models.InstagramData
				var err error
				for i := 0; i < 3; i++ {
					data, err = scraper.ScrapeInstagram(l)
					if err == nil && data != nil {
						break
					}
				}
				if data != nil {
					result = models.TweetData{
						Year:     data.Year,
						Month:    data.Month,
						Day:      data.Day,
						Text:     data.Text,
						Hour:     data.Hour,
						Minute:   data.Minute,
						Title:    data.Title,
						URL:      data.URL,
						Author:   data.Author,
						Likes:    data.Likes,
						Comments: data.Comments,
						Created:  data.Created,
					}
				} else {
					result = models.TweetData{URL: l}
				}
			} else {
				result = models.TweetData{URL: l}
			}

			resultChan <- result
			bar.Add(1)
		}(link)
	}

	wg.Wait()
	close(resultChan)

	var results []models.TweetData
	for res := range resultChan {
		results = append(results, res)
	}
	return results
}
