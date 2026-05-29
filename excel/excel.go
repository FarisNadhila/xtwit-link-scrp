package excel

import (
	"fmt"
	"scrape-neticle-go/models"

	"github.com/xuri/excelize/v2"
)

func ReadLinks(filename string) ([]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(f.GetSheetList()[0])
	if err != nil {
		return nil, err
	}

	var links []string
	linkCol := -1
	if len(rows) > 0 {
		for i, cell := range rows[0] {
			if cell == "Links" {
				linkCol = i
				break
			}
		}
	}

	if linkCol == -1 {
		return nil, fmt.Errorf("'Links' column not found")
	}

	for i := 1; i < len(rows); i++ {
		if len(rows[i]) > linkCol && rows[i][linkCol] != "" {
			links = append(links, rows[i][linkCol])
		}
	}

	return links, nil
}

func WriteResults(templateFile, outputFile string, results []models.TweetData) error {
	f, err := excelize.OpenFile(templateFile)
	if err != nil {
		return err
	}
	defer f.Close()

	sheetName := "Data to upload"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return err
	}

	headerRowIndex := 2
	if len(rows) <= headerRowIndex {
		return fmt.Errorf("template header row not found")
	}

	headers := rows[headerRowIndex]
	colMap := make(map[string]int)

	fields := []string{"Year", "Month", "Day", "Text", "Hour", "Minute", "Title", "URL", "Author", "Views", "Likes"}
	for _, field := range fields {
		for i, h := range headers {
			if h == field {
				colMap[field] = i + 1
				break
			}
		}
	}

	startRow := 4
	for {
		cell, err := f.GetCellValue(sheetName, fmt.Sprintf("A%d", startRow))
		if err != nil || cell == "" {
			break
		}
		startRow++
	}

	for i, res := range results {
		row := startRow + i
		writeData(f, sheetName, row, colMap, res)
	}

	return f.SaveAs(outputFile)
}

func writeData(f *excelize.File, sheet string, row int, colMap map[string]int, data models.TweetData) {
	if c, ok := colMap["Year"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Year)
	}
	if c, ok := colMap["Month"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Month)
	}
	if c, ok := colMap["Day"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Day)
	}
	if c, ok := colMap["Text"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Text)
	}
	if c, ok := colMap["Hour"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Hour)
	}
	if c, ok := colMap["Minute"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Minute)
	}
	if c, ok := colMap["Title"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Title)
	}
	if c, ok := colMap["URL"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.URL)
	}
	if c, ok := colMap["Author"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Author)
	}
	if c, ok := colMap["Views"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Views)
	}
	if c, ok := colMap["Likes"]; ok {
		f.SetCellValue(sheet, cellName(c, row), data.Likes)
	}
}

func cellName(col, row int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return fmt.Sprintf("%s%d", name, row)
}
