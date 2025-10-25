# Bright Data Proxy Demo with Playwright

Web scraping demo using Playwright with Bright Data proxy for browser automation.

## Features

- **Playwright Integration**: Browser automation with Chromium
- **Bright Data Proxy**: Residential proxy support
- **Anti-Detection**: Advanced browser args to mimic real users
- **Web Scraping**: Extract HTML content and take screenshots
- **Environment Variables**: Secure credential management

## Setup

### 1. Create .env file
Create `.env` file in root directory:

```
CUSTOMER_ID=brd-customer-YOUR_ACTUAL_CUSTOMER_ID
PROXY_PASSWORD=YOUR_ACTUAL_ZONE_PASSWORD
```

### 2. Install dependencies
```bash
go mod tidy
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.5200.1 install --with-deps
```

### 3. Run the program
```bash
go run main.go
```

## Environment Variables (Alternative)

Instead of using `.env` file, you can set environment variables directly:

**Windows (PowerShell):**
```powershell
$env:CUSTOMER_ID="brd-customer-YOUR_ACTUAL_CUSTOMER_ID"
$env:PROXY_PASSWORD="YOUR_ACTUAL_ZONE_PASSWORD"
go run main.go
```

**Windows (CMD):**
```cmd
set CUSTOMER_ID=brd-customer-YOUR_ACTUAL_CUSTOMER_ID
set PROXY_PASSWORD=YOUR_ACTUAL_ZONE_PASSWORD
go run main.go
```

**Linux/Mac:**
```bash
export CUSTOMER_ID="brd-customer-YOUR_ACTUAL_CUSTOMER_ID"
export PROXY_PASSWORD="YOUR_ACTUAL_ZONE_PASSWORD"
go run main.go
```

## Output Files

- `ha_content.html` - Scraped HTML content
- `ha_screenshot.png` - Page screenshot

## Anti-Detection Features

The browser is configured with advanced arguments to avoid bot detection:
- Disabled automation indicators
- Realistic user agent
- Standard window size (1920x1080)
- Disabled unnecessary features
- SSL certificate handling
