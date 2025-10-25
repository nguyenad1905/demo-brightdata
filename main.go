// Tên tệp: main.go
package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

// createSessionID tạo một ID ngẫu nhiên (10 ký tự)
// Việc này ra lệnh cho Web Unlocker sử dụng cùng một IP (IP cố định) cho tất cả các yêu cầu
func createSessionID() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func main() {
	log.Println("--- Bắt đầu quy trình đăng nhập không trình duyệt ---")

	// --- 1. TẢI BIẾN MÔI TRƯỜNG TỪ TỆP .ENV ---
	err := godotenv.Load()
	if err != nil {
		log.Println("Không tìm thấy tệp .env, đang sử dụng biến môi trường hệ thống.")
	}

	// --- 2. LẤY THÔNG TIN XÁC THỰC TỪ ENV ---
	CUSTOMER_ID := os.Getenv("BRD_CUSTOMER_ID")
	PROXY_PASSWORD := os.Getenv("BRD_PASSWORD")
	ZONE_NAME := os.Getenv("BRD_ZONE_NAME")
	PROXY_PORT := os.Getenv("BRD_PROXY_PORT")
	MY_USERNAME := os.Getenv("APP_USERNAME")
	MY_PASSWORD := os.Getenv("APP_PASSWORD")

	// Kiểm tra xem tất cả các biến đã được đặt chưa
	if CUSTOMER_ID == "" || PROXY_PASSWORD == "" || MY_USERNAME == "" || MY_PASSWORD == "" || ZONE_NAME == "" || PROXY_PORT == "" {
		log.Fatal("Lỗi: Một trong các biến môi trường (BRD_CUSTOMER_ID, BRD_PASSWORD, BRD_ZONE_NAME, BRD_PROXY_PORT, APP_USERNAME, APP_PASSWORD) chưa được đặt.")
	}

	// --- 3. TẠO HTTP CLIENT (PHIÊN LÀM VIỆC) ---

	// Jar vẫn cần thiết để TỰ ĐỘNG nhận và lưu cookie từ Bước 5 (POST)
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Lỗi khi tạo cookie jar: %v", err)
	}

	sessionID := createSessionID()
	log.Printf("Đang sử dụng Session ID cho IP cố định: %s\n", sessionID)

	// Xây dựng chuỗi username, bao gồm cả session cố định và quốc gia
	proxyUser := fmt.Sprintf("%s-zone-%s-session-%s-brd-country-us", CUSTOMER_ID, ZONE_NAME, sessionID)
	proxyHost := "brd.superproxy.io"

	// Sử dụng PROXY_PORT từ .env
	proxyString := fmt.Sprintf("http://%s:%s@%s:%s", proxyUser, PROXY_PASSWORD, proxyHost, PROXY_PORT)

	proxyURL, err := url.Parse(proxyString)
	if err != nil {
		log.Fatalf("Lỗi khi phân tích URL proxy: %v", err)
	}

	// Cấu hình Transport để dùng proxy và bỏ qua xác minh SSL
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Tạo Client (phiên làm việc) cuối cùng
	client := &http.Client{
		Transport: transport,
		Jar:       jar,               // Dùng để TỰ ĐỘNG LƯU cookie
		Timeout:   120 * time.Second, // Đặt thời gian chờ 2 phút
	}

	// --- 4. BƯỚC 1: GET ĐỂ LẤY TOKEN ---
	loginURL := "https://www.ha.com/c/login.zx"
	log.Printf("Đang gửi GET tới %s (Web Unlocker sẽ giải CAPTCHA)...\n", loginURL)

	responseGet, err := client.Get(loginURL)
	if err != nil {
		log.Fatalf("Lỗi khi gửi yêu cầu GET: %v", err)
	}
	defer responseGet.Body.Close()

	if responseGet.StatusCode != 200 {
		log.Fatalf("Yêu cầu GET thất bại, trạng thái: %s", responseGet.Status)
	}

	// --- 5. BƯỚC 2: PHÂN TÍCH TOKEN ---
	log.Println("Đang phân tích HTML để tìm 'formToken'...")

	doc, err := goquery.NewDocumentFromReader(responseGet.Body)
	if err != nil {
		log.Fatalf("Lỗi khi phân tích HTML: %v", err)
	}

	tokenElement := doc.Find("input[name='formToken']")
	dynamicFormToken, exists := tokenElement.Attr("value")

	if !exists || dynamicFormToken == "" {
		log.Fatal("LỖI: Không tìm thấy 'formToken' trong HTML. Trang web có thể đã thay đổi.")
	}
	log.Printf("Đã tìm thấy formToken động: %s\n", dynamicFormToken)

	// --- 6. BƯỚC 3: GỬI POST ĐỂ ĐĂNG NHẬP ---

	// Xây dựng Form Data (payload) chính xác như trong DevTools
	loginPayload := url.Values{
		"validCheck":          {"valid"},
		"source":              {"nav"},
		"forceLogin":          {""},
		"loginAction":         {"log-in"},
		"formToken":           {dynamicFormToken}, // <-- Sử dụng token động
		"findMe":              {""},
		"username":            {MY_USERNAME}, // <-- Từ .env
		"password":            {MY_PASSWORD}, // <-- Từ .env
		"chkRememberPassword": {"1"},
		"loginButton":         {"Sign in"},
	}

	log.Println("Đang gửi POST với payload (Form Data) để đăng nhập...")

	responsePost, err := client.PostForm(loginURL, loginPayload)
	if err != nil {
		log.Fatalf("Lỗi khi gửi yêu cầu POST: %v", err)
	}
	defer responsePost.Body.Close()

	// 302 (Chuyển hướng) cũng là dấu hiệu đăng nhập thành công
	if responsePost.StatusCode != 200 && responsePost.StatusCode != 302 {
		log.Fatalf("Đăng nhập thất bại, trạng thái: %s", responsePost.Status)
	}

	log.Println("✅ ĐĂNG NHẬP THÀNH CÔNG!")

	// --- 7. BƯỚC 4: KIỂM TRA PHIÊN (SESSION) THỦ CÔNG ---

	// Lấy URL chúng ta vừa đăng nhập để lấy cookie
	loginURL_parsed, _ := url.Parse(loginURL)
	// Lấy tất cả cookie mà JAR đã lưu cho ha.com
	loginCookies := jar.Cookies(loginURL_parsed)

	if len(loginCookies) == 0 {
		log.Fatal("LỖI: Đã đăng nhập nhưng không tìm thấy cookie nào để gửi đi.")
	}

	// Xây dựng chuỗi "Cookie: " thủ công
	var cookieHeader strings.Builder
	for i, cookie := range loginCookies {
		if i > 0 {
			cookieHeader.WriteString("; ")
		}
		// Thêm "tên=giá trị"
		cookieHeader.WriteString(fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
	}
	log.Println("Đã xây dựng header Cookie thủ công để gửi đến tên miền phụ.")

	jewelryURL := "https://jewelry.ha.com/"
	log.Printf("Đang kiểm tra duy trì đăng nhập tại: %s\n", jewelryURL)

	// Tạo một yêu cầu (request) GET mới
	req, err := http.NewRequest("GET", jewelryURL, nil)
	if err != nil {
		log.Fatalf("Lỗi khi tạo yêu cầu GET thủ công: %v", err)
	}

	// === BƯỚC QUAN TRỌNG: Thêm header Cookie thủ công ===
	req.Header.Set("Cookie", cookieHeader.String())
	// Thêm các header khác để trông giống trình duyệt
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/5.37.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/5.37.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Referer", "https://www.ha.com/") // Giả vờ như ta đến từ trang chủ

	// Gửi yêu cầu bằng client (vẫn dùng proxy và session IP)
	responseJewelry, err := client.Do(req)
	if err != nil {
		log.Fatalf("Lỗi khi truy cập trang jewelry: %v", err)
	}
	defer responseJewelry.Body.Close()

	// Đọc nội dung trang jewelry
	body, err := io.ReadAll(responseJewelry.Body)
	if err != nil {
		log.Fatalf("Lỗi khi đọc nội dung trang jewelry: %v", err)
	}

	os.WriteFile("jewelry_page.html", body, 0644)
	log.Println("Đã lưu nội dung vào 'jewelry_page.html'")

	if strings.Contains(string(body), MY_USERNAME) {
		log.Println("✅ THÀNH CÔNG! Đã duy trì đăng nhập trên tên miền phụ.")
	} else {
		log.Println("❌ LỖI: KHÔNG duy trì đăng nhập. Hãy kiểm tra 'jewelry_page.html'.")
	}
}
