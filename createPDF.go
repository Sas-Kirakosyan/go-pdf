package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jung-kurt/gofpdf"
)

func fetchHTML(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(body), nil
}

// take text from html
func extractDescription(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// Find <div class="description">
	desc := doc.Find("div.description").Text()
	fmt.Println("text--", desc)
	return strings.TrimSpace(desc), nil
}

// add custom text
func processText(desc string) string {
	return desc + "\n\nբողոքարկվել \"Պարեկային ծառայության պետին\""
}

func createPDF(text string, filename string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("NotoSansArm", "", "/fonts/noto-sans-armenian-armenian-600-normal.ttf")
	pdf.SetFont("NotoSansArm", "", 14)
	pdf.AddPage()
	pdf.MultiCell(0, 10, text, "", "", false)

	return pdf.OutputFileAndClose(filename)
}

func readVivPoliceAndCreatePdf() {
	fmt.Println("Hello, Go!")
	url := "https://viv.police.am/decision/85GVKLQKDVVE"

	html, err := fetchHTML(url)
	if err != nil {
		panic(err)
	}
	desc, err := extractDescription(html)
	if err != nil {
		panic(err)
	}

	// 3. Add extra text
	finalText := processText(desc)

	// 4. Generate PDF
	err = createPDF(finalText, "output.pdf")
	if err != nil {
		panic(err)
	}

	fmt.Println("PDF created successfully!")
}
