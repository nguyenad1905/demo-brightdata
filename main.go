package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load .env: %v", err)
	}

	// Get Bright Data credentials
	brightDataAPIKey := os.Getenv("BRIGHT_DATA_API_KEY")
	zoneName := os.Getenv("ZONE_NAME")
	if zoneName == "" {
		zoneName = "web_unlocker1"
	}

	if brightDataAPIKey == "" {
		log.Fatal("Please set BRIGHT_DATA_API_KEY environment variable")
	}

	fmt.Println("=== Bright Data Web Unlocker ===")

	// Target URL to access
	targetURL := "https://www.ha.com/c/login.zx?source=nav"

	// Step 1: Use Web Unlocker API to get unlocked content
	fmt.Printf("Requesting %s via Web Unlocker...\n", targetURL)
	unlockedHTML, err := requestWebUnlocker(brightDataAPIKey, zoneName, targetURL)
	if err != nil {
		log.Fatalf("Failed to unlock URL: %v", err)
	}

	fmt.Printf("Successfully unlocked URL. Content size: %d bytes\n", len(unlockedHTML))

	// Step 2: Save HTML to file
	outputFile := "ha_content.html"
	err = os.WriteFile(outputFile, []byte(unlockedHTML), 0644)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}

	fmt.Printf("Content saved to: %s\n", outputFile)
	fmt.Println("\n=== Demo completed ===")
}

// requestWebUnlocker makes request to Bright Data Web Unlocker API
func requestWebUnlocker(apiKey, zone, targetURL string) (string, error) {
	url := "https://api.brightdata.com/request"

	payloadStr := fmt.Sprintf(`{
        "zone": "%s",
        "url": "%s",
        "format": "raw",
        "method": "GET"
    }`, zone, targetURL)

	payload := strings.NewReader(payloadStr)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	// Log response status
	fmt.Printf("Response Status: %d\n", res.StatusCode)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	fmt.Printf("Response Body Length: %d bytes\n", len(body))

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Response Body: %s\n", string(body))
		return "", fmt.Errorf("API error: %d - %s", res.StatusCode, string(body))
	}

	// Log first 500 chars if content is large
	if len(body) > 0 && len(body) < 1000 {
		fmt.Printf("Response Body: %s\n", string(body))
	} else if len(body) > 0 {
		fmt.Printf("Response Body (first 500 chars): %s...\n", string(body[:500]))
	}

	return string(body), nil
}
