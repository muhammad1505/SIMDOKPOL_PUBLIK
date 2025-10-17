#!/bin/bash
# =================================================================
# Setup Nginx Reverse Proxy untuk SIMDOKPOL
# Mengubah akses dari localhost:8080 menjadi domain lokal
# WAJIB DIJALANKAN MENGGUNAKAN 'sudo'
# =================================================================

# Warna untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Cek hak akses sudo
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}ERROR: Harap jalankan skrip ini menggunakan sudo.${NC}"
  echo "Contoh: sudo bash $0"
  exit 1
fi

clear
echo -e "${BLUE}========================================================${NC}"
echo -e "${BLUE}  Setup Nginx Reverse Proxy untuk SIMDOKPOL${NC}"
echo -e "${BLUE}========================================================${NC}"
echo ""

# Input domain dari pengguna
read -p "Masukkan nama domain (default: simdokpol.local): " VHOST_DOMAIN
VHOST_DOMAIN=${VHOST_DOMAIN:-simdokpol.local}

# Konfigurasi
BACKEND_PORT="8080"
NGINX_CONF_DIR="/etc/nginx"
SITES_AVAILABLE="$NGINX_CONF_DIR/sites-available"
SITES_ENABLED="$NGINX_CONF_DIR/sites-enabled"
VHOST_CONF="$SITES_AVAILABLE/$VHOST_DOMAIN.conf"

echo -e "${CYAN}Domain: ${GREEN}http://$VHOST_DOMAIN${NC}"
echo -e "${CYAN}Backend: ${GREEN}http://localhost:$BACKEND_PORT${NC}"
echo ""

# Menu pilihan
PS3="Masukkan pilihan Anda (1-3): "
options=("Setup Virtual Host dengan Nginx" "Hapus Virtual Host" "Keluar")
select opt in "${options[@]}"
do
    case $opt in
        "Setup Virtual Host dengan Nginx")
            echo ""
            echo -e "${BLUE}=== Memulai Setup Virtual Host ===${NC}"
            echo ""
            
            # ============================================
            # 1. Install Nginx jika belum terinstall
            # ============================================
            echo -e "${BLUE}[1/6] Memeriksa instalasi Nginx...${NC}"
            if ! command -v nginx &> /dev/null; then
                echo -e "${YELLOW}[INFO] Nginx belum terinstall. Menginstall...${NC}"
                pacman -S --noconfirm nginx
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}[OK] Nginx berhasil diinstall.${NC}"
                else
                    echo -e "${RED}[ERROR] Gagal menginstall Nginx.${NC}"
                    exit 1
                fi
            else
                echo -e "${GREEN}[OK] Nginx sudah terinstall.${NC}"
            fi
            echo ""
            
            # ============================================
            # 2. Buat direktori sites-available dan sites-enabled jika belum ada
            # ============================================
            echo -e "${BLUE}[2/6] Menyiapkan struktur direktori Nginx...${NC}"
            mkdir -p "$SITES_AVAILABLE"
            mkdir -p "$SITES_ENABLED"
            
            # Pastikan nginx.conf menginclude sites-enabled
            if ! grep -q "include.*sites-enabled" "$NGINX_CONF_DIR/nginx.conf"; then
                # Backup konfigurasi asli
                cp "$NGINX_CONF_DIR/nginx.conf" "$NGINX_CONF_DIR/nginx.conf.backup"
                
                # Tambahkan include ke dalam block http
                sed -i '/http {/a \    include /etc/nginx/sites-enabled/*;' "$NGINX_CONF_DIR/nginx.conf"
                echo -e "${GREEN}[OK] Konfigurasi nginx.conf telah diupdate.${NC}"
            else
                echo -e "${GREEN}[OK] Konfigurasi sites-enabled sudah ada.${NC}"
            fi
            echo ""
            
            # ============================================
            # 3. Buat konfigurasi virtual host
            # ============================================
            echo -e "${BLUE}[3/6] Membuat konfigurasi virtual host...${NC}"
            
            cat > "$VHOST_CONF" << EOF
# ============================================================
# Konfigurasi Virtual Host untuk SIMDOKPOL
# Domain: $VHOST_DOMAIN
# Backend: localhost:$BACKEND_PORT
# Dibuat: $(date)
# ============================================================

server {
    # Listening pada port 80 untuk HTTP
    listen 80;
    listen [::]:80;
    
    # Nama domain yang akan digunakan
    server_name $VHOST_DOMAIN;
    
    # Logging untuk debugging dan monitoring
    access_log /var/log/nginx/${VHOST_DOMAIN}_access.log;
    error_log /var/log/nginx/${VHOST_DOMAIN}_error.log;
    
    # Client max body size untuk upload file (sesuaikan jika perlu)
    client_max_body_size 50M;
    
    # Root location - proxy semua request ke aplikasi Go
    location / {
        # Proxy pass ke backend aplikasi Go di port $BACKEND_PORT
        proxy_pass http://127.0.0.1:$BACKEND_PORT;
        
        # Headers penting untuk preserve informasi client
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # Timeout settings untuk long-running requests
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffer settings untuk performance
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        proxy_busy_buffers_size 8k;
        
        # WebSocket support (jika diperlukan di masa depan)
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # Static files caching (optional - untuk optimasi)
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://127.0.0.1:$BACKEND_PORT;
        proxy_set_header Host \$host;
        
        # Cache static files di browser selama 30 hari
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
    
    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    # Disable akses ke hidden files
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }
}

# ============================================================
# Optional: Redirect www ke non-www
# ============================================================
server {
    listen 80;
    listen [::]:80;
    server_name www.$VHOST_DOMAIN;
    return 301 http://$VHOST_DOMAIN\$request_uri;
}
EOF
            
            echo -e "${GREEN}[OK] File konfigurasi dibuat: $VHOST_CONF${NC}"
            echo ""
            
            # ============================================
            # 4. Aktifkan virtual host dengan symbolic link
            # ============================================
            echo -e "${BLUE}[4/6] Mengaktifkan virtual host...${NC}"
            
            # Hapus symlink lama jika ada
            if [ -L "$SITES_ENABLED/$VHOST_DOMAIN.conf" ]; then
                rm "$SITES_ENABLED/$VHOST_DOMAIN.conf"
            fi
            
            # Buat symbolic link
            ln -s "$VHOST_CONF" "$SITES_ENABLED/$VHOST_DOMAIN.conf"
            echo -e "${GREEN}[OK] Virtual host diaktifkan.${NC}"
            echo ""
            
            # ============================================
            # 5. Tambahkan entri ke /etc/hosts
            # ============================================
            echo -e "${BLUE}[5/6] Menambahkan entri ke /etc/hosts...${NC}"
            HOSTS_ENTRY="127.0.0.1 $VHOST_DOMAIN"
            
            if ! grep -q "$VHOST_DOMAIN" /etc/hosts; then
                # Backup hosts file
                cp /etc/hosts /etc/hosts.backup.$(date +%Y%m%d_%H%M%S)
                
                # Tambahkan entri baru
                echo "$HOSTS_ENTRY" >> /etc/hosts
                echo -e "${GREEN}[OK] Entri ditambahkan ke /etc/hosts${NC}"
            else
                echo -e "${YELLOW}[INFO] Entri sudah ada di /etc/hosts${NC}"
            fi
            echo ""
            
            # ============================================
            # 6. Test konfigurasi dan restart Nginx
            # ============================================
            echo -e "${BLUE}[6/6] Menguji dan me-restart Nginx...${NC}"
            
            # Test konfigurasi Nginx
            nginx -t
            
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}[OK] Konfigurasi Nginx valid.${NC}"
                
                # Restart Nginx
                systemctl restart nginx
                
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}[OK] Nginx berhasil di-restart.${NC}"
                    
                    # Enable Nginx untuk auto-start saat boot
                    systemctl enable nginx
                    echo -e "${GREEN}[OK] Nginx dienable untuk auto-start.${NC}"
                else
                    echo -e "${RED}[ERROR] Gagal me-restart Nginx.${NC}"
                    exit 1
                fi
            else
                echo -e "${RED}[ERROR] Konfigurasi Nginx tidak valid!${NC}"
                echo "Periksa error di atas dan perbaiki konfigurasi."
                exit 1
            fi
            
            echo ""
            echo -e "${GREEN}========================================================${NC}"
            echo -e "${GREEN}  ✓ SETUP SELESAI DENGAN SUKSES!${NC}"
            echo -e "${GREEN}========================================================${NC}"
            echo ""
            echo -e "${CYAN}Informasi Akses:${NC}"
            echo -e "  • Domain Lokal    : ${GREEN}http://$VHOST_DOMAIN${NC}"
            echo -e "  • Backend Server  : ${YELLOW}http://localhost:$BACKEND_PORT${NC}"
            echo -e "  • Log Access      : /var/log/nginx/${VHOST_DOMAIN}_access.log"
            echo -e "  • Log Error       : /var/log/nginx/${VHOST_DOMAIN}_error.log"
            echo ""
            echo -e "${YELLOW}Catatan Penting:${NC}"
            echo "  1. Pastikan aplikasi SIMDOKPOL berjalan di port $BACKEND_PORT"
            echo "  2. Akses aplikasi melalui: ${GREEN}http://$VHOST_DOMAIN${NC}"
            echo "  3. Tidak perlu lagi mengetik port :8080"
            echo "  4. Nginx akan otomatis start setiap boot"
            echo ""
            
            # Test koneksi
            echo -e "${BLUE}[TEST] Mencoba koneksi ke backend...${NC}"
            sleep 2
            
            if command -v curl &> /dev/null; then
                HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$BACKEND_PORT 2>/dev/null)
                
                if [[ "$HTTP_CODE" =~ ^(200|301|302)$ ]]; then
                    echo -e "${GREEN}[OK] Backend aplikasi merespons dengan baik (HTTP $HTTP_CODE)${NC}"
                    echo ""
                    echo -e "${GREEN}✓ Anda sekarang dapat mengakses: http://$VHOST_DOMAIN${NC}"
                else
                    echo -e "${RED}[WARN] Backend tidak merespons (HTTP $HTTP_CODE)${NC}"
                    echo -e "${YELLOW}      Pastikan aplikasi SIMDOKPOL sedang berjalan di port $BACKEND_PORT${NC}"
                fi
            fi
            
            echo ""
            break
            ;;
            
        "Hapus Virtual Host")
            echo ""
            echo -e "${BLUE}=== Menghapus Virtual Host ===${NC}"
            echo ""
            
            # 1. Nonaktifkan virtual host
            echo -e "${BLUE}[1/4] Menonaktifkan virtual host...${NC}"
            if [ -L "$SITES_ENABLED/$VHOST_DOMAIN.conf" ]; then
                rm "$SITES_ENABLED/$VHOST_DOMAIN.conf"
                echo -e "${GREEN}[OK] Virtual host dinonaktifkan.${NC}"
            else
                echo -e "${YELLOW}[INFO] Virtual host tidak aktif.${NC}"
            fi
            
            # 2. Hapus file konfigurasi
            echo -e "${BLUE}[2/4] Menghapus file konfigurasi...${NC}"
            if [ -f "$VHOST_CONF" ]; then
                rm "$VHOST_CONF"
                echo -e "${GREEN}[OK] File konfigurasi dihapus.${NC}"
            else
                echo -e "${YELLOW}[INFO] File konfigurasi tidak ditemukan.${NC}"
            fi
            
            # 3. Hapus entri dari /etc/hosts
            echo -e "${BLUE}[3/4] Menghapus entri dari /etc/hosts...${NC}"
            if grep -q "$VHOST_DOMAIN" /etc/hosts; then
                # Backup dulu
                cp /etc/hosts /etc/hosts.backup.$(date +%Y%m%d_%H%M%S)
                
                # Hapus entri
                sed -i.tmp "/$VHOST_DOMAIN/d" /etc/hosts
                rm -f /etc/hosts.tmp
                echo -e "${GREEN}[OK] Entri dihapus dari /etc/hosts.${NC}"
            else
                echo -e "${YELLOW}[INFO] Entri tidak ditemukan di /etc/hosts.${NC}"
            fi
            
            # 4. Restart Nginx
            echo -e "${BLUE}[4/4] Me-restart Nginx...${NC}"
            systemctl restart nginx
            
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}[OK] Nginx berhasil di-restart.${NC}"
            else
                echo -e "${RED}[ERROR] Gagal me-restart Nginx.${NC}"
            fi
            
            echo ""
            echo -e "${GREEN}========================================================${NC}"
            echo -e "${GREEN}  ✓ PENGHAPUSAN SELESAI${NC}"
            echo -e "${GREEN}========================================================${NC}"
            echo ""
            echo -e "${YELLOW}Akses aplikasi kembali melalui: http://localhost:$BACKEND_PORT${NC}"
            echo ""
            break
            ;;
            
        "Keluar")
            echo ""
            echo -e "${CYAN}Terima kasih!${NC}"
            break
            ;;
            
        *) 
            echo -e "${RED}Pilihan tidak valid: $REPLY${NC}"
            ;;
    esac
done

echo ""