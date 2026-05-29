# Digivla X/Twitter link scraper

## Made with
- Golang 1.26.3
- Excelize (data)
- Goquery (scrape)

## How it works
- Scrape __INITIAL_STATE__ from parsed HTML
- Extract entity to struct
- Return in data slices
- Goroutines the scrape func with 50 buffer channel
- Output to excel via excelize

## How to run
1. On terminal install the lib: go mod tidy
2. Modify the InputFile (excel) in main.go as an input data
3. Make sure the template is in the same folder on project folder
4. Run the code from terminal: go run main.go
5. Or to compile it from terminal use: go build main.go


