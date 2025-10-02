package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/jung-kurt/gofpdf"
)

func readRoadPolice() error {
	pinCode := "25YIE3UJHBFC"
	url := "https://offense.roadpolice.am/violation?pin=" + pinCode

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// give enough time for manual captcha solve if needed
	ctx, cancel = context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	var outerHTML string
	var recaptchaVal string
	var screenshotBuf []byte
	var nodesCount int
	var evalResult string

	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.WaitReady("body", chromedp.ByQuery),

		// Debug: how many buttons
		chromedp.EvaluateAsDevTools(`document.querySelectorAll("button.blue-btn").length`, &nodesCount),

		// Save HTML
		// chromedp.OuterHTML("html", &outerHTML, chromedp.ByQuery),

		// Take screenshot
		chromedp.CaptureScreenshot(&screenshotBuf),

		// Fill PIN
		chromedp.WaitVisible(`#pin`, chromedp.ByQuery),
		chromedp.Focus(`#pin`, chromedp.ByQuery),
		chromedp.SetValue(`#pin`, pinCode, chromedp.ByQuery),

		chromedp.Sleep(800 * time.Millisecond),

		// Try native click
		chromedp.Click(`button.blue-btn`, chromedp.ByQuery),
		chromedp.Sleep(1500 * time.Millisecond),

		// Read recaptcha response
		chromedp.Value(`#recaptchaResponse`, &recaptchaVal, chromedp.ByQuery),
		chromedp.WaitVisible(`#result-inner`, chromedp.ByID),
		chromedp.Text(`#result-inner`, &outerHTML, chromedp.ByID),
	}
	if err := chromedp.Run(ctx, tasks); err != nil {
		_ = os.WriteFile("page_error.html", []byte(outerHTML), 0644)
		return fmt.Errorf("chromedp run failed: %w", err)
	}

	// Save debug artifacts
	if err := os.WriteFile("page.html", []byte(outerHTML), 0644); err != nil {
		log.Printf("failed saving html: %v", err)
	}
	if err := os.WriteFile("page.png", screenshotBuf, 0644); err != nil {
		log.Printf("failed saving screenshot: %v", err)
	}

	// Use gofpdf to write the extracted HTML/text into a PDF

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.AddUTF8Font("NotoSansArm", "", "NotoSansArmenian-Regular.ttf")

	pdf.SetFont("NotoSansArm", "", 14)

	// Add plain text version of HTML
	pdf.MultiCell(0, 10, outerHTML, "", "", false)

	// Save PDF
	if err := pdf.OutputFileAndClose("page.pdf"); err != nil {
		return fmt.Errorf("failed saving pdf: %w", err)
	}

	fmt.Println("button nodes found:", nodesCount)
	fmt.Println("last JS eval result:", evalResult)
	fmt.Println("recaptcha token length:", len(recaptchaVal))
	fmt.Println("page saved to page.html and page.png")

	return nil
}
