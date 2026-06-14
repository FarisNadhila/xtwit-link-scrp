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

type ScrapeTask struct {
	Index int
	URL   string
}

type bucket struct {
	platform platform
	tasks    []ScrapeTask
}

type ScraperManager struct {
	results []models.SocialData
	mu      sync.Mutex
}

func main() {
	inputFile := flag.String("input", "input.xlsx", "Path to the input Excel file")
	outputFile := flag.String("output", "output.xlsx", "Path to the output Excel file")
	flag.Parse()

	buckets, total := analyzeInput(*inputFile)
	manager := &ScraperManager{
		results: make([]models.SocialData, total),
	}
	manager.runParallelScrape(buckets)
	manager.exportToExcel(*outputFile)
}

// analyze
func analyzeInput(filename string) ([]bucket, int) {
	links, err := excel.ReadLinks(filename)
	if err != nil {
		log.Fatalf("Error reading links from %s: %v", filename, err)
	}

	buckets := make([]bucket, len(platforms))
	for i, p := range platforms {
		buckets[i] = bucket{platform: p}
	}

	for idx, link := range links {
		lower := strings.ToLower(link)
		matched := false
		for i, p := range platforms {
			if p.Matches(lower) {
				buckets[i].tasks = append(buckets[i].tasks, ScrapeTask{Index: idx, URL: link})
				matched = true
				break
			}
		}
		if !matched {
			fmt.Printf("Unknown link type skipped: %s\n", link)
		}
	}

	for _, b := range buckets {
		fmt.Printf("%s(%d) ", b.platform.Label, len(b.tasks))
	}
	fmt.Print("\n\n")

	return buckets, len(links)
}

// scrpMgmt
func (m *ScraperManager) runParallelScrape(buckets []bucket) {
	var wg sync.WaitGroup
	for _, b := range buckets {
		if len(b.tasks) == 0 {
			continue
		}
		wg.Add(1)
		go func(b bucket) {
			defer wg.Done()
			bar := progressbar.Default(int64(len(b.tasks)), b.platform.Label)
			m.workerPool(b.tasks, bar, b.platform.Scrape)
		}(b)
	}
	wg.Wait()
}

func (m *ScraperManager) workerPool(tasks []ScrapeTask, bar *progressbar.ProgressBar, scrapeFn func(string) (models.SocialData, error)) {
	taskChan := make(chan ScrapeTask, WorkerCount)

	numWorkers := WorkerCount
	if len(tasks) < numWorkers {
		numWorkers = len(tasks)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskChan {
				data := scrapeWithRetry(task.URL, scrapeFn)
				m.mu.Lock()
				m.results[task.Index] = data
				m.mu.Unlock()
				bar.Add(1)
			}
		}()
	}

	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)
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
			// exponential backoff: 2s, 4s, 8s, ...
			backoff := time.Duration(1<<uint(attempt+1)) * time.Second
			// jitter: rand 0-1000ms
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}
	log.Printf("Failed to scrape %s after %d attempts: %v", link, MaxRetries, err)
	return models.SocialData{URL: link, Text: "FAILED TO SCRAPE"}
}

// excel
func (m *ScraperManager) exportToExcel(outputFile string) {
	var finalResults []models.SocialData
	for _, res := range m.results {
		if res.URL != "" {
			finalResults = append(finalResults, res)
		}
	}

	if len(finalResults) == 0 {
		fmt.Println("No results to export.")
		return
	}
	fmt.Printf("%d results -> %s...\n", len(finalResults), outputFile)
	if err := excel.WriteResults(TemplateFile, outputFile, finalResults); err != nil {
		log.Fatalf("Error writing results to excel: %v", err)
	}
}
