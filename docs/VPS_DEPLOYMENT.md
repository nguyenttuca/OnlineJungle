# Hướng dẫn Deploy Top OJ lên VPS (Ubuntu 22.04+)

Tài liệu này hướng dẫn chi tiết từng bước (step-by-step) cách đưa mã nguồn Top OJ lên một máy chủ ảo (VPS) mới tinh chạy hệ điều hành Ubuntu, cấu hình tên miền (Domain) và SSL (HTTPS).

## 1. Chuẩn bị VPS & Cài đặt phần mềm cơ bản

Sau khi thuê VPS (DigitalOcean, AWS, Vultr, v.v.), đăng nhập vào VPS thông qua SSH:
```bash
ssh root@<IP_CUA_VPS>
```

Cập nhật hệ thống và cài đặt các phần mềm cần thiết:
```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl git build-essential nginx certbot python3-certbot-nginx
```

Cài đặt Golang (phiên bản 1.21 trở lên):
```bash
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
go version # Đảm bảo đã nhận lệnh go
```

## 2. Cài đặt và cấu hình PostgreSQL

Cài đặt PostgreSQL:
```bash
sudo apt install -y postgresql postgresql-contrib
```

Khởi tạo Database và User cho ứng dụng:
```bash
sudo -u postgres psql

# (Trong màn hình psql, chạy các lệnh sau)
CREATE DATABASE oj;
CREATE USER oj_admin WITH ENCRYPTED PASSWORD 'mat_khau_bao_mat_cua_ban';
GRANT ALL PRIVILEGES ON DATABASE oj TO oj_admin;
\q
```

## 3. Clone Mã nguồn và Thiết lập môi trường

Khuyến nghị tạo một user riêng thay vì chạy app bằng `root`:
```bash
adduser --disabled-password --gecos "" topoj
su - topoj
```

Clone mã nguồn:
```bash
git clone <link-github-cua-ban> /home/topoj/oj-web
cd /home/topoj/oj-web
```

Biên dịch (Build) ứng dụng:
```bash
go mod tidy
go build -o top_oj_server ./cmd/server
```

*(Tùy chọn)* Cấu hình biến môi trường bằng cách tạo file `.env` hoặc để Systemd quản lý biến môi trường.

**Khởi tạo Database Schema:**
Bạn cần import các file `.sql` chứa cấu trúc bảng vào database `oj`.
```bash
# Ví dụ nếu bạn có file schema.sql
psql -U oj_admin -d oj -h 127.0.0.1 -W < schema.sql
```

Bấm `Ctrl + D` để thoát khỏi user `topoj` và quay lại `root`.

## 4. Thiết lập Systemd để chạy ngầm (Background Service)

Tạo file cấu hình service cho Top OJ:
```bash
sudo nano /etc/systemd/system/topoj.service
```

Dán nội dung sau vào:
```ini
[Unit]
Description=Top OJ Web Server
After=network.target postgresql.service

[Service]
Type=simple
User=topoj
Group=topoj
WorkingDirectory=/home/topoj/oj-web
ExecStart=/home/topoj/oj-web/top_oj_server
Restart=always
RestartSec=3

# Biến môi trường
Environment="PORT=8080"
Environment="DATABASE_URL=postgres://oj_admin:mat_khau_bao_mat_cua_ban@localhost:5432/oj?sslmode=disable"

[Install]
WantedBy=multi-user.target
```

Kích hoạt và khởi chạy service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable topoj
sudo systemctl start topoj
sudo systemctl status topoj # Kiểm tra xem đang chạy xanh (active) chưa
```

## 5. Cấu hình Nginx (Reverse Proxy) và Tên miền

Đảm bảo bạn đã trỏ Tên miền (A Record) từ nhà cung cấp Domain về `IP_CUA_VPS`. Ví dụ: `oj.yourdomain.com`.

Tạo file cấu hình Nginx:
```bash
sudo nano /etc/nginx/sites-available/topoj
```

Dán nội dung sau (thay `oj.yourdomain.com` bằng tên miền thực tế):
```nginx
server {
    listen 80;
    server_name oj.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Phục vụ file tĩnh trực tiếp qua Nginx để tăng tốc
    location /static/ {
        alias /home/topoj/oj-web/static/;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }
}
```

Kích hoạt cấu hình:
```bash
sudo ln -s /etc/nginx/sites-available/topoj /etc/nginx/sites-enabled/
sudo nginx -t # Kiểm tra xem có lỗi syntax không
sudo systemctl restart nginx
```

## 6. Cài đặt SSL (HTTPS) với Let's Encrypt

Chạy Certbot để tự động đăng ký và cấu hình chứng chỉ SSL miễn phí:
```bash
sudo certbot --nginx -d oj.yourdomain.com
```

Certbot sẽ hỏi bạn một số thông tin (Email, đồng ý điều khoản). Sau đó nó sẽ tự động sửa file cấu hình Nginx của bạn để ép HTTP chuyển hướng sang HTTPS.

## 7. Cấu hình Judge Nodes (Máy chấm)

Hệ thống Web Server đã chạy, nhưng bạn cần cấu hình các "Máy chấm" (Judge Nodes) để code của học sinh có thể được biên dịch và chạy thử.
1. Đăng nhập vào Top OJ bằng tài khoản Admin.
2. Truy cập trang Quản trị (Admin Panel).
3. Thêm URL của các máy chấm (ví dụ: `http://127.0.0.1:25170` hoặc IP của một máy chủ chấm riêng biệt) vào bảng `judge_nodes`. Đảm bảo máy chấm đã được start (thường là một service riêng viết bằng C++ hoặc Python/Go).

## 8. Các lệnh bảo trì thường dùng

- **Xem log hệ thống OJ (Real-time):**
  ```bash
  sudo journalctl -u topoj -f
  ```
- **Khởi động lại OJ (Sau khi update code):**
  ```bash
  sudo systemctl restart topoj
  ```
- **Update mã nguồn mới nhất:**
  ```bash
  su - topoj
  cd ~/oj-web
  git pull origin main
  go build -o top_oj_server ./cmd/server
  exit
  sudo systemctl restart topoj
  ```
