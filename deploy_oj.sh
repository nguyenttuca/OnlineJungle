#!/bin/bash
# Script Auto Deploy cho Top OJ (Ubuntu/Debian)
# Yêu cầu chạy bằng quyền root (sudo)

set -e

# --- CẤU HÌNH ---
DB_NAME="oj"
DB_USER="oj_admin"
DB_PASS="oj_secure_password_123" # Đổi mật khẩu này trên production!
DOMAIN="your-domain.com"
APP_DIR="/var/www/onlinejungle"
PORT=8080
BINARY_NAME="top_oj_server"
# ----------------

echo "🚀 Bắt đầu quá trình Auto Deploy Top OJ..."

if [ "$EUID" -ne 0 ]; then
  echo "❌ Vui lòng chạy script này với quyền root (thêm sudo ở đầu)"
  exit 1
fi

echo "📦 1. Cài đặt các gói cần thiết (Nginx, PostgreSQL, Go)..."
apt-get update
apt-get install -y nginx postgresql postgresql-contrib curl git

# Cài đặt Go nếu chưa có
if ! command -v go &> /dev/null; then
    echo "⬇️ Đang cài đặt Golang 1.22..."
    curl -OL https://golang.org/dl/go1.22.4.linux-amd64.tar.gz
    tar -C /usr/local -xvf go1.22.4.linux-amd64.tar.gz
    rm go1.22.4.linux-amd64.tar.gz
    echo "export PATH=$PATH:/usr/local/go/bin" >> /etc/profile
    source /etc/profile
fi
export PATH=$PATH:/usr/local/go/bin

echo "🗄️ 2. Khởi tạo Cơ sở dữ liệu PostgreSQL..."
sudo -u postgres psql -c "SELECT 1 FROM pg_database WHERE datname = '${DB_NAME}'" | grep -q 1 || sudo -u postgres psql -c "CREATE DATABASE ${DB_NAME};"
sudo -u postgres psql -c "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'" | grep -q 1 || sudo -u postgres psql -c "CREATE USER ${DB_USER} WITH ENCRYPTED PASSWORD '${DB_PASS}';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE ${DB_NAME} TO ${DB_USER};"
sudo -u postgres psql -c "ALTER DATABASE ${DB_NAME} OWNER TO ${DB_USER};"

echo "⚙️ 3. Chuẩn bị mã nguồn và Build..."
# Tạo thư mục app nếu chưa có
mkdir -p ${APP_DIR}
cp -r ./* ${APP_DIR}/

cd ${APP_DIR}
echo "⬇️ Đang tải dependencies và Build..."
go mod tidy
go build -o ${BINARY_NAME} ./cmd/server

# Chạy Database Migration (nếu có thư mục sql/schema)
if [ -d "sql/schema" ]; then
    echo "🔄 Đang chạy DB Migrations..."
    # Tạm dùng lệnh cat đổ thẳng SQL vào DB cho đơn giản
    for f in $(ls sql/schema/*.up.sql | sort); do
        sudo -u postgres psql -d ${DB_NAME} -f "$f" || true
    done
fi

echo "🛡️ 4. Cấu hình Nginx..."
cat > /etc/nginx/sites-available/top_oj <<EOF
server {
    listen 80;
    server_name ${DOMAIN};

    location / {
        proxy_pass http://localhost:${PORT};
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /static/ {
        alias ${APP_DIR}/static/;
        expires 30d;
    }
}
EOF

ln -sf /etc/nginx/sites-available/top_oj /etc/nginx/sites-enabled/
# Xóa default nginx config
rm -f /etc/nginx/sites-enabled/default
systemctl restart nginx

echo "🔄 5. Cấu hình Systemd (Background Service)..."
cat > /etc/systemd/system/topoj.service <<EOF
[Unit]
Description=Top OJ Web Server
After=network.target postgresql.service

[Service]
Type=simple
User=root
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/${BINARY_NAME}
Restart=always
RestartSec=3
Environment="DATABASE_URL=postgres://${DB_USER}:${DB_PASS}@localhost:5432/${DB_NAME}?sslmode=disable"
Environment="PORT=${PORT}"

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable topoj
systemctl restart topoj

echo "✅ HOÀN TẤT! Hệ thống đã được Deploy thành công."
echo "🌍 Hãy trỏ tên miền ${DOMAIN} về IP của server và truy cập để kiểm tra."
echo "📜 Để xem log server chạy ngầm, dùng lệnh: journalctl -u topoj -f"
