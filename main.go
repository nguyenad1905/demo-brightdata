// Tên tệp: browser.go
// Giải pháp này sử dụng Bright Data Scraping Browser (tăng timeout cho Fill)
package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time" // Thêm thư viện time để dùng Sleep

	"github.com/joho/godotenv"
	"github.com/playwright-community/playwright-go"
)

func main() {
	log.Println("--- Bắt đầu quy trình đăng nhập bằng Scraping Browser ---")

	// --- 1. TẢI BIẾN MÔI TRƯỜNG ---
	err := godotenv.Load()
	if err != nil {
		log.Println("Không tìm thấy tệp .env, đang sử dụng biến môi trường hệ thống.")
	}

	CUSTOMER_ID := os.Getenv("BRD_CUSTOMER_ID")
	PROXY_PASSWORD := os.Getenv("BRD_PASSWORD")
	ZONE_NAME := os.Getenv("BRD_SB_ZONE_NAME") // Zone của Scraping Browser
	SB_PORT := os.Getenv("BRD_SB_PORT")        // Cổng của Scraping Browser (ví dụ: 9222)
	MY_USERNAME := os.Getenv("APP_USERNAME")
	MY_PASSWORD := os.Getenv("APP_PASSWORD")

	// Kiểm tra các biến quan trọng cho mã này
	if CUSTOMER_ID == "" || PROXY_PASSWORD == "" || ZONE_NAME == "" || SB_PORT == "" || MY_USERNAME == "" || MY_PASSWORD == "" {
		log.Fatal("Lỗi: Một trong các biến môi trường (BRD_CUSTOMER_ID, BRD_PASSWORD, BRD_SB_ZONE_NAME, BRD_SB_PORT, APP_USERNAME, APP_PASSWORD) chưa được đặt.")
	}

	// --- 2. KẾT NỐI VỚI SCRAPING BROWSER CỦA BRIGHT DATA ---

	// Xây dựng URL điểm cuối (endpoint) WebSocket (wss)
	endpointURL := fmt.Sprintf("wss://%s-zone-%s:%s@brd.superproxy.io:%s", CUSTOMER_ID, ZONE_NAME, PROXY_PASSWORD, SB_PORT)

	log.Println("Đang kết nối với:", fmt.Sprintf("wss://%s-zone-%s:***PASSWORD***@brd.superproxy.io:%s", CUSTOMER_ID, ZONE_NAME, SB_PORT))

	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Không thể khởi động Playwright: %v", err)
	}
	defer pw.Stop()

	// Kết nối với trình duyệt từ xa của Bright Data
	browser, err := pw.Chromium.ConnectOverCDP(endpointURL, playwright.BrowserTypeConnectOverCDPOptions{
		Timeout: playwright.Float(120000), // Tăng thời gian chờ kết nối lên 2 phút
	})
	if err != nil {
		log.Fatalf("Không thể kết nối với Scraping Browser: %v", err)
	}
	defer browser.Close()

	log.Println("Đã kết nối! Đang tạo bối cảnh (context)...")
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

	// --- 3. THỰC HIỆN ĐĂNG NHẬP (Mã Playwright đơn giản) ---
	loginURL := "https://www.ha.com/c/login.zx"

	// BƯỚC 1: Tải trang.
	log.Printf("Đang tải trang %s (Scraping Browser sẽ giải CAPTCHA)...\n", loginURL)
	if _, err := page.Goto(loginURL, playwright.PageGotoOptions{Timeout: playwright.Float(120000)}); err != nil {
		log.Fatalf("Lỗi khi tải trang login: %v", err)
	}

	// Thêm độ trễ để Scraping Browser có thời gian xử lý CAPTCHA (nếu cần)
	log.Println("Chờ 5 giây để Scraping Browser xử lý (nếu có)...")
	time.Sleep(5 * time.Second)

	// Chụp ảnh màn hình để xem trang đã tải như thế nào
	log.Println("Trang đã tải, đang chụp ảnh màn hình trước khi điền...")
	if _, err := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String("debug_login_page.png"), FullPage: playwright.Bool(true)}); err != nil {
		log.Printf("Không thể chụp ảnh màn hình: %v", err)
	}
	log.Println("Đã lưu ảnh chụp màn hình vào 'debug_login_page.png'")

	// === Selector (Giả định selector password là 'input#password') ===
	usernameSelector := "input[name='username']"
	passwordSelector := "input#password" // Sử dụng ID

	log.Printf("Đang chờ selector '%s' xuất hiện...\n", usernameSelector)
	if _, err := page.WaitForSelector(usernameSelector, playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(10000)}); err != nil {
		log.Fatalf("LỖI: Không tìm thấy '%s' trong 10s!", usernameSelector)
	}
	log.Printf("Đang chờ selector '%s' xuất hiện...\n", passwordSelector)
	if _, err := page.WaitForSelector(passwordSelector, playwright.PageWaitForSelectorOptions{Timeout: playwright.Float(10000)}); err != nil {
		log.Fatalf("LỖI: Không tìm thấy '%s' trong 10s!", passwordSelector)
	}
	// =============================================================

	// BƯỚC 2: Điền thông tin
	log.Println("Đang điền thông tin đăng nhập...")
	if err := page.Fill(usernameSelector, MY_USERNAME); err != nil { // Timeout mặc định 30s
		log.Fatalf("Không thể điền username (%s): %v", usernameSelector, err)
	}

	// === TĂNG TIMEOUT CHO FILL PASSWORD ===
	log.Println("Đang điền password (timeout 60s)...")
	if err := page.Fill(passwordSelector, MY_PASSWORD, playwright.PageFillOptions{
		Timeout: playwright.Float(60000), // Tăng lên 60 giây
	}); err != nil {
		log.Fatalf("Không thể điền password (%s): %v", passwordSelector, err)
	}
	// ===================================

	// BƯỚC 3: Nhấp "Sign in"
	log.Println("Đang nhấp nút đăng nhập...")
	// Selector cho nút đăng nhập
	loginButtonSelector := "button[name='loginButton']"
	if err := page.Click(loginButtonSelector); err != nil {
		log.Fatalf("Không thể nhấp nút login (%s): %v", loginButtonSelector, err)
	}

	// Chờ trang tải sau khi nhấp đăng nhập
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: playwright.Float(60000),
	}); err != nil {
		log.Println("Lỗi khi chờ trang sau đăng nhập, nhưng vẫn tiếp tục kiểm tra...")
	}

	log.Println("✅ ĐÃ ĐĂNG NHẬP THÀNH CÔNG (Giả định)!")

	// --- 4. KIỂM TRA TÊN MIỀN PHỤ ---
	jewelryURL := "https://jewelry.ha.com/"
	log.Printf("Đang kiểm tra duy trì đăng nhập tại: %s\n", jewelryURL)

	if _, err := page.Goto(jewelryURL, playwright.PageGotoOptions{Timeout: playwright.Float(120000)}); err != nil {
		log.Fatalf("Lỗi khi tải trang jewelry: %v", err)
	}

	// Lấy nội dung trang
	content, err := page.Content()
	if err != nil {
		log.Fatalf("Không thể lấy nội dung trang jewelry: %v", err)
	}
	os.WriteFile("jewelry_page_sb.html", []byte(content), 0644)
	log.Println("Đã lưu nội dung vào 'jewelry_page_sb.html'")

	// Kiểm tra bằng tên người dùng hoặc các chuỗi khác
	if strings.Contains(content, MY_USERNAME) || strings.Contains(content, "Welcome") || strings.Contains(content, "Sign-Out") {
		log.Println("✅✅✅ THÀNH CÔNG! Đã duy trì đăng nhập trên tên miền phụ.")
	} else {
		log.Println("❌ LỖI: KHÔNG duy trì đăng nhập. Hãy kiểm tra 'jewelry_page_sb.html'.")
	}
}
