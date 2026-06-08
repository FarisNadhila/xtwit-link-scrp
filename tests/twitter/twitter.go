package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"time"
	// "github.com/PuerkitoBio/goquery"
)

func main() {
	var link string = "https://x.com/libranazero/status/2059239292501700724"

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

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	os.WriteFile("tests/twitter/response.html", bodyBytes, 0644)

	// doc, err := goquery.NewDocumentFromReader(resp.Body)
	// if err != nil {
	// 	log.Fatal("error")
	// }
}
