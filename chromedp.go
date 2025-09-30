package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/chromedp/chromedp"
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
		chromedp.OuterHTML("html", &outerHTML, chromedp.ByQuery),

		// Take screenshot
		chromedp.CaptureScreenshot(&screenshotBuf),

		// Fill PIN
		chromedp.WaitVisible(`#pin`, chromedp.ByQuery),
		chromedp.Focus(`#pin`, chromedp.ByQuery),
		chromedp.SetValue(`#pin`, pinCode, chromedp.ByQuery),

		chromedp.Sleep(800 * time.Millisecond),

		// Try native click
		chromedp.Click(`button.blue-btn`, chromedp.ByQuery),

		chromedp.Sleep(800 * time.Millisecond),

		// Try JS click
		chromedp.EvaluateAsDevTools(`(function(){
            const btn = document.querySelector("button.blue-btn");
            if(!btn) return "no-button";
            btn.scrollIntoView();
            if (typeof btn.click === "function") { btn.click(); return "clicked-js"; }
            return "no-click-fn";
        })()`, &evalResult),

		chromedp.Sleep(1200 * time.Millisecond),

		// Try form submit
		chromedp.EvaluateAsDevTools(`(function(){
            const f = document.querySelector("#form-block");
            if(!f) return "no-form";
            try { f.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true })); return "dispatched-submit"; } 
            catch(e) { try { f.submit(); return "native-submit"; } catch(e2){ return "submit-failed"; } }
        })()`, &evalResult),

		chromedp.Sleep(1500 * time.Millisecond),

		// Read recaptcha response
		chromedp.Value(`#recaptchaResponse`, &recaptchaVal, chromedp.ByQuery),
	}

	if err := chromedp.Run(ctx, tasks); err != nil {
		_ = ioutil.WriteFile("page_error.html", []byte(outerHTML), 0644)
		return fmt.Errorf("chromedp run failed: %w", err)
	}

	// Save debug artifacts
	if err := ioutil.WriteFile("page.html", []byte(outerHTML), 0644); err != nil {
		log.Printf("failed saving html: %v", err)
	}
	if err := ioutil.WriteFile("page.png", screenshotBuf, 0644); err != nil {
		log.Printf("failed saving screenshot: %v", err)
	}

	fmt.Println("button nodes found:", nodesCount)
	fmt.Println("last JS eval result:", evalResult)
	fmt.Println("recaptcha token length:", len(recaptchaVal))
	fmt.Println("page saved to page.html and page.png")

	return nil
}
