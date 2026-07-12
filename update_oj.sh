#!/bin/bash
# Script Auto Update cho Top OJ (Ubuntu/Debian)
# Chạy bằng quyền root (sudo)

set -e

APP_DIR="/var/www/onlinejungle"
BINARY_NAME="top_oj_server"

echo "🔄 Bắt đầu quá trình Auto Update Top OJ..."

if [ "$EUID" -ne 0 ]; then
  echo "❌ Vui lòng chạy script này với quyền root (thêm sudo ở đầu)"
  exit 1
fi

if [ ! -d "$APP_DIR/.git" ]; then
  echo "❌ Thư mục $APP_DIR không phải là một Git repository hoặc không tồn tại."
  echo "⚠️ Script này chỉ hoạt động nếu mã nguồn được clone trực tiếp bằng Git trên server."
  exit 1
fi

cd $APP_DIR

echo "⬇️ Đang kéo code mới nhất từ GitHub (git pull)..."
# Lưu ý: Cần đảm bảo server VPS có cấu hình xác thực (SSH key hoặc PAT) với GitHub.
git pull origin main

echo "🔨 Đang cài đặt dependencies và Build..."
export PATH=$PATH:/usr/local/go/bin
go mod tidy
go build -o $BINARY_NAME ./cmd/server

echo "🔄 Đang chạy database migrations (nếu có)..."
if [ -d "sql/schema" ]; then
    for f in $(ls sql/schema/*.up.sql | sort); do
        sudo -u postgres psql -d oj -f "$f" || true
    done
fi

echo "🔄 Đang khởi động lại dịch vụ Top OJ..."
systemctl restart topoj

echo "✅ HOÀN TẤT! Hệ thống đã được Update thành công."
echo "📜 Để xem log server chạy ngầm, dùng lệnh: journalctl -u topoj -f"
