# X/Twt link scraper

## Made with
- Golang 1.26.3
- Excelize (data)
- Goquery (scrape)

## How it works
- Scrape `__INITIAL_STATE__` from parsed HTML
- Extract entity to struct
- Return in data slices
- Goroutines the scrape func with 50 buffer channel
- Output to excel via excelize

## How to run
1. On terminal install the lib:
```bash
go mod tidy
```
2. Prepare your input Excel file (default is `input.xlsx`)
3. Make sure the template (`template.xlsx`) is in the same folder as the project folder
4. **Run** the code from terminal:
```bash
go run main.go
```
   Or specify a custom input file:
   ```bash
   go run main.go -input input.xlsx
   ```
5. Or to **compile** it to an executable:
   - For Linux: 
   ```bash
   go build -o xtwit-link-scrp
   ```
   - For Windows:
   ```bash
   OOS=windows GOARCH=amd64 go build -o xtwit-link-scrp.exe
   ```
   - Run the executable:
   ```bash
   ./xtwit-link-scrp -input my_links.xlsx
   ```

## ADD-ON (for threads & tiktok)
Tiktok and Threads using advance anti-bot system to prevent data scraping.
It need some workaround to bypass the rate-limiting/anti-bot.
To make the scraping process works makes sure to install:
- chromedp lib: on terminal ->
```bash
GOPROXY=direct GONOSUMDB=* go get github.com/chromedp/chromedp
```
- Chrome/Chromium browser. `chrome canary` recommended.
