#!/bin/bash

# Qigent One-Click Deployment Script
# Usage: sudo ./deploy_server.sh

set -e # Exit immediately if a command exits with a non-zero status.

# --- Configuration ---
APP_DIR="/opt/qigent"
BACKEND_PORT="8090"
SERVER_IP="38.147.185.103"

# Colors
GREEN='\033[0;32m'
NC='\033[0m'

echo -e "${GREEN}Starting Deployment...${NC}"

# 1. Install Dependencies (Ubuntu/Debian)
echo -e "${GREEN}[1/6] Installing dependencies...${NC}"
apt-get update
# Suppress prompts
DEBIAN_FRONTEND=noninteractive apt-get install -y git nginx golang curl mysql-server

# Install Node.js 22.x (Required by Vite)
if ! node -v | grep -q "v22"; then
    echo "Installing Node.js 22.x..."
    curl -fsSL https://deb.nodesource.com/setup_22.x -o nodesource_setup.sh
    bash nodesource_setup.sh
    DEBIAN_FRONTEND=noninteractive apt-get install -y nodejs
    rm nodesource_setup.sh
fi

# 2. Setup Directory
echo -e "${GREEN}[2/6] Setting up directory...${NC}"
mkdir -p "$APP_DIR"
# Copy contents from current dir (assuming script is run from repo)
cp -r . "$APP_DIR"
cd "$APP_DIR" || exit

# 3. Build Backend
echo -e "${GREEN}[3/6] Building Backend...${NC}"
export GOOS=linux
export GOARCH=amd64

# Cleanup old conflicting files (caused by SCP merge)
rm -f internal/api/conversation.go
rm -f internal/api/config.go

go build -o qigent-server main.go

# 4. Build Frontend
echo -e "${GREEN}[4/6] Building Frontend...${NC}"
cd frontend
# Ensure dependencies are installed
rm -rf node_modules
npm install

# Generate .env.production to ensure correct API URL
echo "VITE_API_BASE_URL=http://$SERVER_IP:$BACKEND_PORT" > .env.production

npm run build
cd ..

# 5. Configure Backend Service (Systemd)
echo -e "${GREEN}[5/6] Configuring Backend Service (Systemd)...${NC}"

# Kill any existing manual instances
pkill qigent-server || true

# Configure MySQL (Idempotent)
echo "Configuring MySQL..."
# Create Database
mysql -u root -e "CREATE DATABASE IF NOT EXISTS qigent;" || echo "DB creation skipped (maybe requires password?)"
# Create User (qigent / qigent_secret) - Adjust if needed
mysql -u root -e "CREATE USER IF NOT EXISTS 'qigent'@'localhost' IDENTIFIED BY 'qigent_secret';" || true
mysql -u root -e "GRANT ALL PRIVILEGES ON qigent.* TO 'qigent'@'localhost';" || true
mysql -u root -e "FLUSH PRIVILEGES;" || true

cat > /etc/systemd/system/qigent.service <<EOF
[Unit]
Description=Qigent Backend API
After=network.target mysql.service

[Service]
Type=simple
User=root
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/qigent-server
Restart=on-failure
Environment="PORT=$BACKEND_PORT"
Environment="DSN=qigent:qigent_secret@tcp(127.0.0.1:3306)/qigent?charset=utf8mb4&parseTime=True&loc=Local"
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable qigent
systemctl restart qigent

echo -e "Backend status:"
systemctl status qigent --no-pager | head -n 10

# 6. Configure Nginx
echo -e "${GREEN}[6/6] Configuring Nginx...${NC}"
rm -f /etc/nginx/sites-enabled/default

cat > /etc/nginx/sites-available/qigent <<EOF
server {
    listen 80;
    server_name $SERVER_IP;

    root $APP_DIR/frontend/dist;
    index index.html;

    # Frontend Static Files
    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
EOF

ln -sf /etc/nginx/sites-available/qigent /etc/nginx/sites-enabled/
nginx -t && systemctl restart nginx

echo -e "${GREEN}Deployment Complete!${NC}"
echo -e "Frontend: http://$SERVER_IP"
echo -e "Backend:  http://$SERVER_IP:$BACKEND_PORT"
