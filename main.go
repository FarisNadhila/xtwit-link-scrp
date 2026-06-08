package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
	"xtwit-link-scrp/excel"
	"xtwit-link-scrp/models"
	"xtwit-link-scrp/scraper"

	"github.com/schollz/progressbar/v3"
)

// const
const (
	TemplateFile = "template.xlsx"
	OutputFile   = "output.xlsx"
	MaxRetries   = 3
	WorkerCount  = 20
)

// Platform registry
type platform struct {
	Label   string
	Matches func(lowercaseURL string) bool
	Scrape  func(url string) (models.SocialData, error)
}

var platforms = []platform{
	{
		Label:   "Twitter",
		Matches: func(l string) bool { return strings.Contains(l, "twitter.com") || strings.Contains(l, "x.com") },
		Scrape: func(url string) (models.SocialData, error) {
			data, err := scraper.ScrapeTwitter(url)
			if err != nil {
				return models.SocialData{}, err
			}
			return *data, nil
		},
	},
	{
		Label:   "Instagram",
		Matches: func(l string) bool { return strings.Contains(l, "instagram.com") },
		Scrape: func(url string) (models.SocialData, error) {
			data, err := scraper.ScrapeInstagram(url)
			if err != nil {
				return models.SocialData{}, err
			}
			return *data, nil
		},
	},
	{
		Label:   "TikTok",
		Matches: func(l string) bool { return strings.Contains(l, "tiktok.com") },
		Scrape: func(url string) (models.SocialData, error) {
			data, err := scraper.ScrapeTiktok(url)
			if err != nil {
				return models.SocialData{}, err
			}
			return *data, nil
		},
	},
	{
		Label:   "Threads",
		Matches: func(l string) bool { return strings.Contains(l, "threads.com") },
		Scrape: func(url string) (models.SocialData, error) {
			data, err := scraper.ScrapeThreads(url)
			if err != nil {
				return models.SocialData{}, err
			}
			return *data, nil
		},
	},
}

type bucket struct {
	platform platform
	links    []string
}

type ScraperManager struct {
	results []models.SocialData
	mu      sync.Mutex
}

func main() {
	inputFile := flag.String("input", "input.xlsx", "Path to the input Excel file")
	flag.Parse()

	buckets := analyzeInput(*inputFile)
	manager := &ScraperManager{}
	manager.runParallelScrape(buckets)
	manager.exportToExcel()
}

// analyze
func analyzeInput(filename string) []bucket {
	links, err := excel.ReadLinks(filename)
	if err != nil {
		log.Fatalf("Error reading links from %s: %v", filename, err)
	}

	buckets := make([]bucket, len(platforms))
	for i, p := range platforms {
		buckets[i] = bucket{platform: p}
	}

	for _, link := range links {
		lower := strings.ToLower(link)
		matched := false
		for i, p := range platforms {
			if p.Matches(lower) {
				buckets[i].links = append(buckets[i].links, link)
				matched = true
				break
			}
		}
		if !matched {
			fmt.Printf("Unknown link type skipped: %s\n", link)
		}
	}

	for _, b := range buckets {
		fmt.Printf("%s(%d) ", b.platform.Label, len(b.links))
	}
	fmt.Print("\n\n")

	return buckets
}

// scrpMgmt
func (m *ScraperManager) runParallelScrape(buckets []bucket) {
	var wg sync.WaitGroup
	for _, b := range buckets {
		if len(b.links) == 0 {
			continue
		}
		wg.Add(1)
		go func(b bucket) {
			defer wg.Done()
			bar := progressbar.Default(int64(len(b.links)), b.platform.Label)
			m.workerPool(b.links, bar, b.platform.Scrape)
		}(b)
	}
	wg.Wait()
}

func (m *ScraperManager) workerPool(links []string, bar *progressbar.ProgressBar, scrapeFn func(string) (models.SocialData, error)) {
	linkChan := make(chan string, WorkerCount)

	numWorkers := WorkerCount
	if len(links) < numWorkers {
		numWorkers = len(links)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range linkChan {
				data := scrapeWithRetry(link, scrapeFn)
				m.mu.Lock()
				m.results = append(m.results, data)
				m.mu.Unlock()
				bar.Add(1)
			}
		}()
	}

	for _, link := range links {
		linkChan <- link
	}
	close(linkChan)
	wg.Wait()
}

func scrapeWithRetry(link string, scrapeFn func(string) (models.SocialData, error)) models.SocialData {
	var (
		data models.SocialData
		err  error
	)
	for attempt := 0; attempt < MaxRetries; attempt++ {
		data, err = scrapeFn(link)
		if err == nil {
			return data
		}

		if attempt < MaxRetries-1 {
			// Exponential backoff: 2s, 4s, 8s...
			backoff := time.Duration(1<<uint(attempt+1)) * time.Second
			// Jitter: add random 0-1000ms
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}
	log.Printf("Failed to scrape %s after %d attempts: %v", link, MaxRetries, err)
	return models.SocialData{URL: link, Text: "FAILED TO SCRAPE"}
}

// excel
func (m *ScraperManager) exportToExcel() {
	if len(m.results) == 0 {
		fmt.Println("No results to export.")
		return
	}
	fmt.Printf("%d results -> %s...\n", len(m.results), OutputFile)
	if err := excel.WriteResults(TemplateFile, OutputFile, m.results); err != nil {
		log.Fatalf("Error writing results to excel: %v", err)
	}
}
