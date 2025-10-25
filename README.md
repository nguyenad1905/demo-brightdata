# Bright Data Proxy Demo with Scrapling

Web scraping demo using Scrapling with Bright Data proxy and Camoufox browser for advanced anti-detection.

## Features

- **Scrapling Integration**: Advanced web scraping library with anti-detection
- **Camoufox Browser**: Stealth browser designed to avoid detection
- **Bright Data Proxy**: Residential proxy support
- **Cloudflare Solver**: Auto-solve Cloudflare challenges
- **Anti-Detection**: Advanced stealth settings to mimic real users
- **Web Scraping**: Extract HTML content and take screenshots
- **Environment Variables**: Secure credential management

## Quick Start

### 1. Set environment variables
**Windows (CMD):**
```cmd
set CUSTOMER_ID=brd-customer-YOUR_ACTUAL_CUSTOMER_ID
set PROXY_PASSWORD=YOUR_ACTUAL_ZONE_PASSWORD
```

**Windows (PowerShell):**
```powershell
$env:CUSTOMER_ID="brd-customer-YOUR_ACTUAL_CUSTOMER_ID"
$env:PROXY_PASSWORD="YOUR_ACTUAL_ZONE_PASSWORD"
```

### 2. Install dependencies
```bash
pip install camoufox python-dotenv
```

### 3. Run the program
```bash
python main.py
```

**Target URL:** `https://www.ha.com/c/login.zx?source=nav`


## Output Files

- `ha_content.html` - Scraped HTML content
- `ha_screenshot.png` - Page screenshot

## Anti-Detection Features

Scrapling with Camoufox provides advanced anti-detection capabilities:
- **Camoufox Browser**: Stealth browser designed to avoid detection
- **Cloudflare Solver**: Auto-solve Cloudflare challenges
- **Stealth Mode**: Advanced stealth settings to mimic real users
- **Proxy Support**: Bright Data residential proxy integration
- **Human-like Behavior**: Simulated mouse movements and scrolling
- **Realistic User Agent**: Chrome 120.0.0.0 user agent
- **Standard Viewport**: 1920x1080 resolution
- **Disabled Automation Indicators**: Removes webdriver properties
