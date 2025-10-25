// Tên tệp: browser.go
// Thử dùng Focus + Keyboard.InsertText cho password
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time" // Thêm thư viện time

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

func main() {
	// ... (Tất cả mã tải .env và kết nối giữ nguyên) ...
	log.Println("--- Bắt đầu quy trình đăng nhập bằng Scraping Browser ---")
	err := godotenv.Load()
	if err != nil {
		log.Println("Không tìm thấy tệp .env...")
	}

	CUSTOMER_ID := os.Getenv("BRD_CUSTOMER_ID")
	PROXY_PASSWORD := os.Getenv("BRD_PASSWORD")
	ZONE_NAME := os.Getenv("BRD_SB_ZONE_NAME")
	SB_PORT := os.Getenv("BRD_SB_PORT")
	MY_USERNAME := os.Getenv("APP_USERNAME")
	MY_PASSWORD := os.Getenv("APP_PASSWORD")

	if CUSTOMER_ID == "" || PROXY_PASSWORD == "" || ZONE_NAME == "" || SB_PORT == "" || MY_USERNAME == "" || MY_PASSWORD == "" {
		log.Fatal("Lỗi: Biến môi trường chưa được đặt.")
	}

	endpointURL := fmt.Sprintf("wss://%s-zone-%s:%s@brd.superproxy.io:%s", CUSTOMER_ID, ZONE_NAME, PROXY_PASSWORD, SB_PORT)
	log.Println("Đang kết nối với:", fmt.Sprintf("wss://%s-zone-%s:***PASSWORD***@brd.superproxy.io:%s", CUSTOMER_ID, ZONE_NAME, SB_PORT))

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Không thể khởi động Playwright: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.ConnectOverCDP(endpointURL, playwright.BrowserTypeConnectOverCDPOptions{
		Timeout: playwright.Float(120000),
	})
	if err != nil {
		log.Fatalf("Không thể kết nối với Scraping Browser: %v", err)
	}
	defer browser.Close()

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36"),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
	})
	if err != nil {
		log.Fatalf("Không thể tạo bối cảnh: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("Không thể tạo trang: %v", err)
	}
	// Không cần SetViewportSize

	// --- 3. THỰC HIỆN ĐĂNG NHẬP ---
	loginURL := "https://www.ha.com/c/login.zx"

	log.Printf("Đang tải trang %s (Scraping Browser sẽ giải CAPTCHA)...\n", loginURL)
	if _, err := page.Goto(loginURL, playwright.PageGotoOptions{Timeout: playwright.Float(120000)}); err != nil {
		log.Fatalf("Lỗi khi tải trang login: %v", err)
	}

	log.Println("Chờ 5 giây để Scraping Browser xử lý (nếu có)...")
	time.Sleep(5 * time.Second)

	log.Println("Trang đã tải, đang chụp ảnh màn hình trước khi điền...")
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String("debug_login_page_before_fill.png"), FullPage: playwright.Bool(true)}); err != nil {
		log.Printf("Không thể chụp ảnh màn hình: %v", err)
	}
	log.Println("Đã lưu ảnh chụp màn hình vào 'debug_login_page_before_fill.png'")

	usernameSelector := "input[name='username']"
	passwordSelector := "input#password"

	log.Printf("Đang chờ selector '%s' xuất hiện...\n", usernameSelector)
	usernameLocator := page.Locator(usernameSelector)
	if err := usernameLocator.WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(10000)}); err != nil {
		log.Fatalf("LỖI: Không tìm thấy '%s' trong 10s!", usernameSelector)
	}
	log.Printf("Đang chờ selector '%s' xuất hiện...\n", passwordSelector)
	passwordLocator := page.Locator(passwordSelector)
	if err := passwordLocator.WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(10000)}); err != nil {
		log.Fatalf("LỖI: Không tìm thấy '%s' trong 10s!", passwordSelector)
	}

	// === BƯỚC 2.1: ĐIỀN USERNAME (Dùng Fill cho đơn giản) ===
	log.Println("Đang điền username...")
	if err := usernameLocator.Fill(MY_USERNAME); err != nil {
		log.Fatalf("Không thể điền username (%s): %v", usernameSelector, err)
	}
	log.Println("Đã điền xong username.")
	time.Sleep(500 * time.Millisecond) // Chờ nửa giây

	// === BƯỚC 2.2: ĐIỀN PASSWORD BẰNG FOCUS + INSERTTEXT ===
	log.Println("Đang thử focus vào ô password...")
	if err := passwordLocator.Focus(playwright.LocatorFocusOptions{Timeout: playwright.Float(10000)}); err != nil {
		// Nếu không focus được thì khả năng cao InsertText cũng thất bại
		log.Fatalf("Không thể focus vào ô password (%s): %v", passwordSelector, err)
	}
	log.Println("Focus thành công.")
	time.Sleep(500 * time.Millisecond) // Chờ nửa giây sau khi focus

	log.Println("Đang thử điền password bằng Keyboard.InsertText...")
	if err := page.Keyboard().InsertText(MY_PASSWORD); err != nil {
		log.Fatalf("Không thể điền password bằng Keyboard.InsertText: %v", err)
	}
	log.Println("Đã điền password bằng Keyboard.InsertText.")
	// ===================================================

	// BƯỚC 3: Nhấp "Sign in"
	log.Println("Đang nhấp nút đăng nhập...")
	if err := page.Click("button[name='loginButton']"); err != nil {
		log.Fatalf("Không thể nhấp nút login: %v", err)
	}

	// ... (Phần còn lại của mã kiểm tra tên miền phụ giữ nguyên) ...
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(60000),
	}); err != nil {
		log.Println("Lỗi khi chờ trang đăng nhập/điều hướng, nhưng vẫn tiếp tục kiểm tra...")
	}
	log.Println("✅ ĐÃ ĐĂNG NHẬP THÀNH CÔNG (Giả định)!")

	// --- 4. KIỂM TRA TÊN MIỀN PHỤ ---
	jewelryURL := "https://jewelry.ha.com/"
	log.Printf("Đang kiểm tra duy trì đăng nhập tại: %s\n", jewelryURL)

	if _, err := page.Goto(jewelryURL, playwright.PageGotoOptions{Timeout: playwright.Float(120000)}); err != nil {
		log.Fatalf("Lỗi khi tải trang jewelry: %v", err)
	}

	content, err := page.Content()
	if err != nil {
		log.Fatalf("Không thể lấy nội dung trang jewelry: %v", err)
	}
	os.WriteFile("jewelry_page_sb.html", []byte(content), 0644)
	log.Println("Đã lưu nội dung vào 'jewelry_page_sb.html'")

	if strings.Contains(content, MY_USERNAME) || strings.Contains(content, "My Account") || strings.Contains(content, "Sign Out") {
		log.Println("✅✅✅ THÀNH CÔNG! Đã duy trì đăng nhập trên tên miền phụ.")
	} else {
		log.Println("❌ LỖI: KHÔNG duy trì đăng nhập. Hãy kiểm tra 'jewelry_page_sb.html'.")
	}
}
