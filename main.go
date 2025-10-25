package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Session stores cookies and session info
type Session struct {
	Cookies   string
	SessionID string
}

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

	fmt.Println("=== Bright Data Web Unlocker Login ===")

	// Target URL to access
	loginURL := "https://www.ha.com/c/login.zx"
	loginPageURL := loginURL + "?source=nav"

	// Initialize session with a unique session_id
	// This will be added to the zone name as "-session-<session_id>"
	sessionID := fmt.Sprintf("session_%d", time.Now().UnixNano())
	session := &Session{
		SessionID: sessionID,
	}
	fmt.Printf("Initialized session_id: %s\n", sessionID)

	// Step 1: Use Web Unlocker API to get login page HTML
	fmt.Printf("Step 1: Getting login page from %s...\n", loginPageURL)
	unlockedHTML, err := requestWebUnlocker(brightDataAPIKey, zoneName, loginPageURL, session)
	if err != nil {
		log.Fatalf("Failed to unlock URL: %v", err)
	}
	fmt.Printf("✓ Successfully got login page. Content size: %d bytes\n", len(unlockedHTML))

	// Step 2: Extract formToken from HTML
	formToken := extractFormToken(unlockedHTML)
	if formToken == "" {
		log.Fatal("Failed to extract formToken from login page")
	}
	fmt.Printf("✓ Extracted formToken: %s\n", formToken)

	// Step 3: Submit login form via Web Unlocker POST
	fmt.Println("\nStep 2: Submitting login form...")
	loginResponse, err := submitLoginForm(brightDataAPIKey, zoneName, loginURL, formToken, session)
	if err != nil {
		log.Fatalf("Failed to submit login: %v", err)
	}
	fmt.Printf("✓ Login response received. Content size: %d bytes\n", len(loginResponse))

	// Print session cookies after login
	fmt.Printf("✓ Session Cookies: %s\n", session.Cookies)

	// Save response to file
	outputFile := "login_response.html"
	err = os.WriteFile(outputFile, []byte(loginResponse), 0644)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}
	fmt.Printf("✓ Login response saved to: %s\n", outputFile)

	// Step 3: Access jewelry.ha.com after login
	fmt.Println("\nStep 3: Accessing jewelry.ha.com...")
	jewelryURL := "https://jewelry.ha.com/"
	jewelryHTML, err := requestWebUnlocker(brightDataAPIKey, zoneName, jewelryURL, session)
	if err != nil {
		log.Fatalf("Failed to access jewelry page: %v", err)
	}
	fmt.Printf("✓ Successfully accessed jewelry page. Content size: %d bytes\n", len(jewelryHTML))

	// Save jewelry page to file
	jewelryFile := "jewelry_page.html"
	err = os.WriteFile(jewelryFile, []byte(jewelryHTML), 0644)
	if err != nil {
		log.Fatalf("Failed to save file: %v", err)
	}
	fmt.Printf("✓ Jewelry page saved to: %s\n", jewelryFile)

	fmt.Println("\n=== All steps completed ===")
}

// extractFormToken extracts the formToken from HTML
func extractFormToken(html string) string {
	// Look for <input type="hidden" name="formToken" value="...">
	re := regexp.MustCompile(`<input[^>]+name="formToken"[^>]+value="([^"]+)"`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// submitLoginForm submits login form via Web Unlocker POST and extracts cookies
func submitLoginForm(apiKey, zone, loginURL, formToken string, session *Session) (string, error) {
	// Create URL-encoded form body as a string
	formBody := fmt.Sprintf("validCheck=valid&source=nav&forceLogin=&loginAction=log-in&formToken=%s&findMe=&username=nguyenad1905c1&password=@!Qwerty3145&chkRememberPassword=1&loginButton=Sign+In",
		formToken)

	// Build headers with cookies if available
	headersJSON := `"Content-Type": "application/x-www-form-urlencoded"`
	if session.Cookies != "" {
		escapedCookies := strings.ReplaceAll(session.Cookies, `\`, `\\`)
		headersJSON = fmt.Sprintf(`"Content-Type": "application/x-www-form-urlencoded",
            "Cookie": %q`, escapedCookies)
		fmt.Printf("Login with cookies: %s...\n", session.Cookies[:100])
	}

	// Create POST request payload for Web Unlocker
	payloadStr := fmt.Sprintf(`{
        "zone": "%s",
        "url": "%s",
        "format": "raw",
        "method": "POST",
        "session": "%s",
        "body": %q,
        "headers": {
            %s
        }
    }`, zone, loginURL, session.SessionID, formBody, headersJSON)

	fmt.Printf("Using session_id: %s\n", session.SessionID)

	fmt.Printf("POST Request to: %s\n", loginURL)
	fmt.Printf("Form Data: username=nguyenad1905c1, formToken=%s...\n", formToken[:8])

	payload := strings.NewReader(payloadStr)

	req, err := http.NewRequest("POST", "https://api.brightdata.com/request", payload)
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

	fmt.Printf("Response Status: %d\n", res.StatusCode)

	// Extract cookies from response headers
	extractCookiesFromResponse(res, session)

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
		fmt.Printf("Response Preview (first 500 chars): %s...\n", string(body[:500]))
	}

	return string(body), nil
}

// requestWebUnlocker makes request to Bright Data Web Unlocker API with session support
func requestWebUnlocker(apiKey, zone, targetURL string, session *Session) (string, error) {
	url := "https://api.brightdata.com/request"

	// Build payload with cookies if available
	payloadStr := fmt.Sprintf(`{
        "zone": "%s",
        "url": "%s",
        "format": "raw",
        "method": "GET",
        "session": "%s"
    }`, zone, targetURL, session.SessionID)

	// Add cookies to headers if available
	if session.Cookies != "" {
		// Escape backslashes in cookies for JSON
		escapedCookies := strings.ReplaceAll(session.Cookies, `\`, `\\`)
		payloadStr = fmt.Sprintf(`{
        "zone": "%s",
        "url": "%s",
        "format": "raw",
        "method": "GET",
        "session": "%s",
        "headers": {
            "Cookie": %q
        }
    }`, zone, targetURL, session.SessionID, escapedCookies)
		fmt.Printf("Sending cookies: %s\n", session.Cookies[:100]+"...")
	}

	fmt.Printf("Using session_id: %s\n", session.SessionID)

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

	// Extract cookies from response headers
	extractCookiesFromResponse(res, session)

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

// extractCookiesFromResponse extracts cookies from response headers
func extractCookiesFromResponse(res *http.Response, session *Session) {
	cookies := res.Cookies()
	if len(cookies) > 0 {
		var cookieStrings []string
		for _, cookie := range cookies {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
		session.Cookies = strings.Join(cookieStrings, "; ")
	}
}
