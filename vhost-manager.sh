#!/bin/bash
# =================================================================
# Manajer Virtual Host untuk SIMDOKPOL (Linux/macOS)
# Versi Final dengan Aturan Port Forwarding Permanen
# WAJIB DIJALANKAN MENGGUNAKAN 'sudo'
# =================================================================

# 1. Cek hak akses sudo
if [ "$EUID" -ne 0 ]; then
  echo "ERROR: Harap jalankan skrip ini menggunakan sudo."
  exit
fi

clear
echo "================================================="
echo " Manajer Virtual Host SIMDOKPOL (Mode Offline)"
echo "================================================="
echo ""

# 2. Meminta input domain dari pengguna
read -p "Masukkan nama domain (default: simdokpol.local): " VHOST_DOMAIN
VHOST_DOMAIN=${VHOST_DOMAIN:-simdokpol.local} # Gunakan default jika input kosong

echo "Domain yang akan digunakan: $VHOST_DOMAIN"
echo ""

# 3. Tampilkan menu pilihan
PS3="Masukkan pilihan Anda: "
options=("Setup Virtual Host (Permanen)" "Hapus Virtual Host (Permanen)" "Keluar")
select opt in "${options[@]}"
do
    case $opt in
        "Setup Virtual Host (Permanen)")
            echo ""
            echo "--- Melakukan Setup Virtual Host untuk $VHOST_DOMAIN ---"
            
            # Langkah 1: Instal iptables-persistent untuk menyimpan aturan
            echo "[1/4] Memeriksa paket 'iptables-persistent'..."
            if ! pacman -Qs iptables-persistent > /dev/null; then
                echo "[INFO] Menginstal 'iptables-persistent'..."
                pacman -S --noconfirm iptables-persistent
            else
                echo "[OK] Paket 'iptables-persistent' sudah terinstal."
            fi

            # Langkah 2: Tambahkan entri ke /etc/hosts
            echo "[2/4] Menambahkan entri ke /etc/hosts..."
            HOSTS_ENTRY="127.0.0.1 $VHOST_DOMAIN"
            if ! grep -q "$VHOST_DOMAIN" /etc/hosts; then
              echo "$HOSTS_ENTRY" >> /etc/hosts
              echo "[OK] Entri '$VHOST_DOMAIN' telah ditambahkan."
            else
              echo "[INFO] Entri '$VHOST_DOMAIN' sudah ada."
            fi

            # Langkah 3: Tambahkan aturan Port Forwarding
            echo "[3/4] Menambahkan aturan port forwarding..."
            iptables -t nat -A PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080

            # Langkah 4: Simpan aturan iptables agar permanen
            echo "[4/4] Menyimpan aturan agar permanen..."
            iptables-save > /etc/iptables/iptables.rules
            systemctl enable iptables.service
            echo "[OK] Aturan berhasil disimpan dan akan dimuat saat boot."
            
            echo ""
            echo "--- SETUP SELESAI ---"
            break
            ;;
        "Hapus Virtual Host (Permanen)")
            echo ""
            echo "--- Menghapus Virtual Host untuk $VHOST_DOMAIN ---"

            # Langkah 1: Hapus entri dari /etc/hosts
            echo "[1/3] Menghapus entri dari /etc/hosts..."
            sed -i.bak "/127.0.0.1 $VHOST_DOMAIN/d" /etc/hosts
            echo "[OK] Entri hosts telah dihapus."

            # Langkah 2: Hapus aturan Port Forwarding
            echo "[2/3] Menghapus aturan port forwarding..."
            iptables -t nat -D PREROUTING -p tcp --dport 80 -j REDIRECT --to-port 8080
            
            # Langkah 3: Simpan kembali aturan yang sudah kosong
            echo "[3/3] Menyimpan perubahan aturan..."
            iptables-save > /etc/iptables/iptables.rules
            echo "[OK] Aturan port forwarding telah dihapus secara permanen."
            
            echo ""
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