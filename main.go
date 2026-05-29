package main

import (
	"fmt"
	"log"
	"scrape-neticle-go/excel"
	"scrape-neticle-go/models"
	"scrape-neticle-go/scraper"
	"sync"

	"github.com/schollz/progressbar/v3"
)

const (
	InputFile    = "../Data loss Twitter.xlsx"
	TemplateFile = "../neticle_mention_excel_upload_template (17).xlsx"
	OutputFile   = "scraped_results.xlsx"
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

			var data *models.TweetData
			var err error
			for i := 0; i < 3; i++ {
				data, err = scraper.ScrapeTwitter(l)
				if err == nil && data != nil {
					break
				}
			}

			if data != nil {
				resultChan <- *data
			} else {
				resultChan <- models.TweetData{URL: l}
			}
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
