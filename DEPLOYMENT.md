# Hướng dẫn Deploy và Tech Stack cho Top OJ

Tài liệu này mô tả chi tiết ngăn xếp công nghệ (tech stack), kiến trúc hệ thống và hướng dẫn các bước để deploy ứng dụng Top OJ lên môi trường server.

## 1. Ngăn xếp công nghệ (Tech Stack)

Hệ thống được thiết kế theo mô hình Client-Server kết hợp Worker Pool cho Judge System.

- **Backend (Web Server & API)**:
  - Ngôn ngữ: Go (Golang) 1.21+
  - Web Framework: `chi` (github.com/go-chi/chi/v5) để định tuyến (routing).
  - Middleware: Các middleware custom để quản lý Session (Cookie-based), Authentication (Role: admin/user) và Security (CSP headers, CORS).
  - Rendering: Go `html/template` cho Server-Side Rendering (SSR).

- **Cơ sở dữ liệu (Database)**:
  - Hệ quản trị CSDL: PostgreSQL 15+.
  - Thư viện kết nối: `pgx/v5` (github.com/jackc/pgx/v5).
  - Code Generator: `sqlc` (tự động sinh type-safe Go code từ raw SQL). Cơ chế này giúp tránh SQL Injection và tăng tốc độ phát triển.
  - Quản lý Migration: Khuyến nghị dùng `golang-migrate` (để up/down schema).

- **Kiến trúc chấm bài (Judge System)**:
  - Dispatcher: Chạy goroutine ngầm trong Web Server, polling CSDL mỗi 500ms để lấy bài nộp (status = `queued`).
  - Worker Pool: Web Server quản lý N workers (goroutines) gọi API tới các Judge Nodes qua HTTP.
  - Judge Node: Các server chấm bài độc lập (được cấu hình URL trong bảng `judge_nodes`), nhận mã nguồn, compile và chạy test case trong môi trường cô lập (sandbox), sau đó trả về kết quả từng test case qua API.

- **Frontend**:
  - Giao diện: HTML5, CSS3.
  - CSS Framework: Bootstrap 5 (tải qua CDN).
  - Icons: Bootstrap Icons và FontAwesome (tải qua CDN).
  - Font: Google Fonts.

## 2. Yêu cầu hệ thống (Prerequisites)

Để chạy hệ thống, server của bạn cần được cài đặt sẵn:
- **Go** (phiên bản 1.21 trở lên).
- **PostgreSQL** (chạy local hoặc remote database).
- **sqlc** (công cụ CLI để sinh mã SQL nếu cần chỉnh sửa thêm).
- Ít nhất 1 Node chấm bài (Judge Node) có API tương thích (ví dụ Pika Judge hoặc server tự build).

## 3. Hướng dẫn Deploy (Môi trường Linux/Ubuntu)

### Bước 1: Chuẩn bị Cơ sở dữ liệu
1. Cài đặt PostgreSQL:
   ```bash
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   ```
2. Tạo User và Database cho hệ thống OJ:
   ```bash
   sudo -u postgres psql
   CREATE DATABASE oj;
   CREATE USER oj_admin WITH ENCRYPTED PASSWORD 'mat_khau_cua_ban';
   GRANT ALL PRIVILEGES ON DATABASE oj TO oj_admin;
   \q
   ```
3. Khởi tạo Schema:
   Chạy các file SQL tạo bảng (thường nằm trong `sql/schema/` hoặc file dump) vào database `oj`.

### Bước 2: Clone source code và cấu hình
1. Lấy mã nguồn:
   ```bash
   git clone https://github.com/nguyenttuca/OnlineJungle
   cd OnlineJungle
   ```
2. Cài đặt các gói phụ thuộc (Dependencies):
   ```bash
   go mod tidy
   ```
3. (Tùy chọn) Chạy lệnh sinh code SQL nếu bạn vừa thay đổi các file trong `internal/database/queries`:
   ```bash
   sqlc generate
   ```

### Bước 3: Cấu hình biến môi trường và kết nối DB
Trong file `main.go` hoặc `cmd/server/main.go`, hệ thống kết nối DB qua Data Source Name (DSN). Hãy đảm bảo chuỗi kết nối khớp với môi trường của bạn:
```go
// Ví dụ DSN
dbURL := "postgres://oj_admin:mat_khau_cua_ban@localhost:5432/oj?sslmode=disable"
```
*(Khuyến nghị: Trong production, nên lấy DSN từ biến môi trường `os.Getenv("DATABASE_URL")`)*

### Bước 4: Chạy Server hoặc Build Binary
- **Chạy trực tiếp (Development)**:
  ```bash
  go run ./cmd/server
  ```
- **Build ra file thực thi (Production)**:
  ```bash
  go build -o top_oj_server ./cmd/server
  ./top_oj_server
  ```
Lúc này server sẽ chạy ở port `:8080` (mặc định).

### Bước 5: Cấu hình Reverse Proxy (Nginx) - Bắt buộc cho Production
Không nên để Web Server Go trực tiếp mở port 80/443 ra internet. Hãy dùng Nginx để proxy:

1. Cài đặt Nginx:
   ```bash
   sudo apt install nginx
   ```
2. Tạo file cấu hình `/etc/nginx/sites-available/top_oj`:
   ```nginx
   server {
       listen 80;
       server_name domain-cua-ban.com;

       location / {
           proxy_pass http://localhost:8080;
           proxy_set_header Host $host;
           proxy_set_header X-Real-IP $remote_addr;
           proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
           proxy_set_header X-Forwarded-Proto $scheme;
       }

       # Tùy chọn cache các static file nếu cần thiết
       location /static/ {
           alias /duong/dan/toi/source/oj-web/static/;
           expires 30d;
       }
   }
   ```
3. Kích hoạt và restart Nginx:
   ```bash
   sudo ln -s /etc/nginx/sites-available/top_oj /etc/nginx/sites-enabled/
   sudo systemctl restart nginx
   ```

### Bước 6: Quản lý Process (Systemd / Supervisor)
Để server Go tự động khởi động lại khi bị crash hoặc khi reboot VPS, hãy tạo Systemd service:

Tạo file `/etc/systemd/system/topoj.service`:
```ini
[Unit]
Description=Top OJ Web Server
After=network.target postgresql.service

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/oj-web
ExecStart=/home/ubuntu/oj-web/top_oj_server
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```
Kích hoạt:
```bash
sudo systemctl daemon-reload
sudo systemctl enable topoj
sudo systemctl start topoj
```

## 4. Danh sách các tính năng cốt lõi
- **Quản lý Problem**: Admin có thể tạo, sửa bài tập.
- **Quản lý Contest**: Tạo kỳ thi với mốc thời gian, thêm các bài tập vào contest.
- **Tính điểm Standings**: Hỗ trợ chấm real-time hai chế độ:
  - **ICPC**: Đếm số bài AC, sử dụng thời gian (penalty minutes) để tie-break.
  - **IOI**: Chỉ lấy điểm số lớn nhất của từng bài, tie-break theo tên bảng chữ cái (không tính thời gian/penalty).
- **Tự động bảo mật CSP**: Đã chặn các script độc hại thực thi trên trình duyệt.
- **Điều phối Worker (Dispatcher)**: Giao bài cho Judge node rảnh, tự cập nhật trạng thái `judging`, `done`, `failed`.
