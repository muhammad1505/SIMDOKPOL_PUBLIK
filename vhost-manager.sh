#!/bin/bash
# =================================================================
# Manajer Virtual Host untuk SIMDOKPOL (Linux/macOS)
# Versi Interaktif dengan Domain Kustom
# WAJIB DIJALANKAN MENGGUNAKAN 'sudo'
# =================================================================

# 1. Cek hak akses sudo
if [ "$EUID" -ne 0 ]; then
  echo "ERROR: Harap jalankan skrip ini menggunakan sudo."
  exit
fi

clear
echo "================================================="
echo " Manajer Virtual Host SIMDOKPOL"
echo "================================================="
echo ""

# 2. Meminta input domain dari pengguna
read -p "Masukkan nama domain (default: simdokpol.local): " VHOST_DOMAIN
VHOST_DOMAIN=${VHOST_DOMAIN:-simdokpol.local} # Gunakan default jika input kosong

echo "Domain yang akan digunakan: $VHOST_DOMAIN"
echo ""

# 3. Tampilkan menu pilihan
PS3="Masukkan pilihan Anda: "
options=("Setup Virtual Host" "Hapus Virtual Host" "Keluar")
select opt in "${options[@]}"
do
    case $opt in
        "Setup Virtual Host")
            echo ""
            echo "--- Melakukan Setup Virtual Host untuk $VHOST_DOMAIN ---"
            
            # Menambahkan entri ke /etc/hosts jika belum ada
            HOSTS_ENTRY="127.0.0.1 $VHOST_DOMAIN"
            if ! grep -q "$HOSTS_ENTRY" /etc/hosts; then
              echo "$HOSTS_ENTRY" >> /etc/hosts
              echo "[OK] Entri '$VHOST_DOMAIN' telah ditambahkan ke /etc/hosts."
            else
              echo "[INFO] Entri '$VHOST_DOMAIN' sudah ada di /etc/hosts."
            fi

            # Mengatur Port Forwarding (Port 80 -> 8080)
            echo "[PROSES] Mengatur port forwarding dari 80 ke 8080..."
            iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
            echo "[OK] Port forwarding berhasil diatur (aturan bersifat sementara hingga reboot)."
            echo "--- SETUP SELESAI ---"
            break
            ;;
        "Hapus Virtual Host")
            echo ""
            echo "--- Menghapus Virtual Host untuk $VHOST_DOMAIN ---"

            # Menghapus entri dari /etc/hosts
            echo "[PROSES] Menghapus entri '$VHOST_DOMAIN' dari /etc/hosts..."
            # Opsi -i.bak membuat backup sebelum mengedit
            sed -i.bak "/127.0.0.1 $VHOST_DOMAIN/d" /etc/hosts
            echo "[OK] Entri hosts telah dihapus."

            # Menghapus aturan Port Forwarding
            echo "[PROSES] Menghapus port forwarding..."
            iptables -t nat -D PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
            echo "[OK] Port forwarding telah dihapus."
            echo "--- PENGHAPUSAN SELESAI ---"
            break
            ;;
        "Keluar")
            break
            ;;
        *) echo "Pilihan tidak valid $REPLY";;
    esac
done

echo ""