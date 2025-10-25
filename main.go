// package main cho biết đây là một chương trình có thể thực thi được.
package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load .env: %v", err)
	}

	// Get proxy credentials
	customerID := os.Getenv("CUSTOMER_ID")
	proxyPassword := os.Getenv("PROXY_PASSWORD")

	if customerID == "" || proxyPassword == "" {
		log.Fatal("Please set environment variables: CUSTOMER_ID, PROXY_PASSWORD")
	}

	fmt.Println("=== Playwright with Bright Data Proxy ===")

	// Playwright with Proxy
	testPlaywright(customerID, proxyPassword)
}

// Playwright test
func testPlaywright(customerID, proxyPassword string) {
	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Failed to start Playwright: %v", err)
	}
	defer pw.Stop()

	// Configure proxy
	proxyServer := fmt.Sprintf("brd.superproxy.io:33335")
	proxyUsername := customerID
	proxyPasswordValue := proxyPassword

	// Launch browser with proxy
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Proxy: &playwright.Proxy{
			Server:   fmt.Sprintf("http://%s", proxyServer),
			Username: &proxyUsername,
			Password: &proxyPasswordValue,
		},
		Headless: playwright.Bool(false),
		Args: []string{
			// SSL and security
			"--ignore-certificate-errors",
			"--ignore-ssl-errors",
			"--disable-web-security",
			"--allow-running-insecure-content",

			// Anti-detection (most important)
			"--disable-blink-features=AutomationControlled",
			"--disable-features=VizDisplayCompositor",
			"--disable-ipc-flooding-protection",
			"--disable-renderer-backgrounding",
			"--disable-backgrounding-occluded-windows",
			"--disable-client-side-phishing-detection",
			"--disable-sync",
			"--disable-translate",
			"--disable-default-apps",
			"--disable-extensions",
			"--disable-plugins",

			// Performance and stability
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--memory-pressure-off",

			// User agent and viewport
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			"--window-size=1920,1080",
		},
	})
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer browser.Close()

	// Create new page
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}

	// Scrape ha.com
	fmt.Println("Scraping ha.com via proxy")
	targetURL := "https://www.ha.com/"
	_, err = page.Goto(targetURL)
	if err != nil {
		log.Printf("Failed to access ha.com: %v", err)
	} else {
		// Wait for page load
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle})

		// Get HTML content
		content, err := page.Content()
		if err != nil {
			log.Printf("Failed to get content: %v", err)
		} else {
			fmt.Printf("Content size: %d bytes\n", len(content))

			// Save HTML
			err = os.WriteFile("ha_content.html", []byte(content), 0644)
			if err != nil {
				log.Printf("Failed to save file: %v", err)
			} else {
				fmt.Println("Saved to: ha_content.html")
			}

			// Extract page info
			fmt.Println("\n--- Page Info ---")

			// Get title
			title, err := page.Locator("h1, title").TextContent()
			if err == nil {
				fmt.Printf("Title: %s\n", strings.TrimSpace(title))
			}

			// Get headings
			headings, err := page.Locator("h1, h2, h3").All()
			if err == nil {
				fmt.Printf("Found %d headings:\n", len(headings))
				for i, heading := range headings {
					if i >= 5 {
						break
					}
					text, err := heading.TextContent()
					if err == nil {
						fmt.Printf("- %s\n", strings.TrimSpace(text))
					}
				}
			}
		}

		// Take screenshot
		screenshot, err := page.Screenshot()
		if err != nil {
			log.Printf("Failed to take screenshot: %v", err)
		} else {
			err = os.WriteFile("ha_screenshot.png", screenshot, 0644)
			if err != nil {
				log.Printf("Failed to save screenshot: %v", err)
			} else {
				fmt.Println("Saved screenshot: ha_screenshot.png")
			}
		}
	}

	fmt.Println("\n=== Demo completed ===")
}
