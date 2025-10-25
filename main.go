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
	context, err := pw.Chromium.LaunchPersistentContext("", playwright.BrowserTypeLaunchPersistentContextOptions{
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

			// Performance and stability
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-gpu",
			"--memory-pressure-off",

			// User agent and viewport
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
			"--window-size=1920,1080",
		},
	})
	if err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer context.Close()

	// Create new page
	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("Failed to create page: %v", err)
	}

	// Add stealth scripts to avoid detection (advanced)
	page.AddInitScript(playwright.Script{
		Content: playwright.String(`
			// Hide webdriver property
			Object.defineProperty(navigator, 'webdriver', {
				get: () => false,
				configurable: true,
			});
			delete navigator.webdriver;
			
			// Mock plugins
			Object.defineProperty(navigator, 'plugins', {
				get: () => [{
					0: {type: 'application/x-google-chrome-pdf', suffixes: 'pdf', description: 'Portable Document Format'},
					1: {type: 'application/pdf', suffixes: 'pdf', description: ''},
					description: 'Portable Document Format',
					filename: 'internal-pdf-viewer',
					length: 2,
					name: 'Chrome PDF Plugin'
				}],
			});
			
			// Mock languages
			Object.defineProperty(navigator, 'languages', {
				get: () => ['en-US', 'en'],
			});
			
			// Mock chrome object (important for Cloudflare detection)
			Object.defineProperty(window, 'chrome', {
				get: () => ({
					runtime: {
						onConnect: undefined,
						onMessage: undefined,
					},
					loadTimes: function() {},
					csi: function() {},
					app: {}
				}),
			});
			
			// Override permissions
			const originalQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) => (
				parameters.name === 'notifications' ?
					Promise.resolve({ state: Notification.permission }) :
					originalQuery(parameters)
			);
			
			// Mock platform
			Object.defineProperty(navigator, 'platform', {
				get: () => 'Win32',
			});
			
			// Override getBattery if it exists
			if (navigator.getBattery) {
				navigator.getBattery = () => Promise.resolve({
					charging: true,
					chargingTime: 0,
					dischargingTime: Infinity,
					level: 1,
				});
			}
			
			// Remove automation indicators
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Array;
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Promise;
			delete window.cdc_adoQpoasnfa76pfcZLmcfl_Symbol;
			
			// Override document properties
			Object.defineProperty(document, 'hidden', {get: () => false});
			Object.defineProperty(document, 'visibilityState', {get: () => 'visible'});
		`),
	})

	// Set extra HTTP headers to look more like a real browser
	err = page.SetExtraHTTPHeaders(map[string]string{
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "en-US,en;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Cache-Control":             "max-age=0",
	})
	if err != nil {
		log.Printf("Failed to set extra HTTP headers: %v", err)
	}

	// Scrape ha.com
	fmt.Println("Scraping ha.com via proxy")

	// First, visit homepage to establish browser fingerprint
	homepageURL := "https://www.ha.com/"
	fmt.Println("Step 1: Visiting homepage to establish browser session...")
	_, err = page.Goto(homepageURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   playwright.Float(30000),
	})
	if err != nil {
		log.Printf("Failed to access homepage: %v", err)
	} else {
		// Simulate human behavior
		mouse := page.Mouse()
		mouse.Move(500, 400)
		page.WaitForTimeout(1000)
		mouse.Move(600, 500)
		page.WaitForTimeout(2000)
	}

	// Now navigate to login page
	targetURL := "https://www.ha.com/c/login.zx?source=nav"
	fmt.Println("Step 2: Navigating to login page...")
	_, err = page.Goto(targetURL, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(60000),
	})
	if err != nil {
		log.Printf("Failed to access ha.com: %v", err)
	} else {
		// Wait a bit for any protection scripts to run
		page.WaitForTimeout(3000)

		// Wait for page to be fully loaded
		page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle})

		// Check if we're being blocked by looking for common block messages
		pageContent, _ := page.Content()
		if strings.Contains(pageContent, "Please enable JS") ||
			strings.Contains(pageContent, "Please disable your ad blocker") ||
			strings.Contains(pageContent, "Checking your browser") {
			log.Println("WARNING: Page appears to be blocking access. Trying alternative approach...")

			// Wait for potential challenge to complete
			page.WaitForTimeout(5000)

			// Try to interact with the page to trigger JS execution
			mouse := page.Mouse()
			mouse.Move(100, 100)
			page.WaitForTimeout(1000)
		}

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
	fmt.Println("Press Enter to close browser...")
	fmt.Scanln() // Wait for user input before closing browser
}
