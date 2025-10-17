@echo off
:: =================================================================
:: Manajer Virtual Host untuk SIMDOKPOL (Windows)
:: Versi Interaktif dengan Domain Kustom
:: WAJIB DIJALANKAN SEBAGAI ADMINISTRATOR
:: =================================================================

:: 1. Cek hak akses administrator
>nul 2>&1 "%SYSTEMROOT%\system32\cacls.exe" "%SYSTEMROOT%\system32\config\system"
if '%errorlevel%' NEQ '0' (
    echo.
    echo  ERROR: Hak Akses Administrator Diperlukan.
    echo  ----------------------------------------------------
    echo  Silakan klik kanan file ini dan pilih 'Run as administrator'.
    echo.
    pause
    exit /b
)

:main_menu
cls
echo =================================================
echo  Manajer Virtual Host SIMDOKPOL
echo =================================================
echo.

:: 2. Meminta input domain dari pengguna
set /p VHOST_DOMAIN="Masukkan nama domain (default: simdokpol.local): "
if "%VHOST_DOMAIN%"=="" set VHOST_DOMAIN=simdokpol.local
echo Domain yang akan digunakan: %VHOST_DOMAIN%
echo.

echo Pilih Aksi:
echo   1. Setup Virtual Host
echo   2. Hapus Virtual Host
echo   3. Keluar
echo.

choice /c 123 /m "Masukkan pilihan Anda (1, 2, atau 3): "

if errorlevel 3 goto :eof
if errorlevel 2 goto :remove_vhost
if errorlevel 1 goto :setup_vhost

:setup_vhost
cls
echo --- Melakukan Setup Virtual Host untuk %VHOST_DOMAIN% ---
echo.

:: Menambahkan entri ke file hosts
set HOSTS_ENTRY="127.0.0.1 %VHOST_DOMAIN%"
findstr /C:%HOSTS_ENTRY% "%SYSTEMROOT%\System32\drivers\etc\hosts" >nul
if %errorlevel% NEQ 0 (
    echo %HOSTS_ENTRY% >> "%SYSTEMROOT%\System32\drivers\etc\hosts"
    echo [OK] Entri '%VHOST_DOMAIN%' telah ditambahkan ke file hosts.
) else (
    echo [INFO] Entri '%VHOST_DOMAIN%' sudah ada di file hosts.
)

:: Mengatur Port Forwarding (Port Proxy)
echo [PROSES] Mengatur port forwarding dari 80 ke 8080...
netsh interface portproxy add v4tov4 listenport=80 listenaddress=127.0.0.1 connectport=8080 connectaddress=127.0.0.1 >nul

echo [OK] Port forwarding berhasil diatur.
echo.
echo --- SETUP SELESAI ---
goto :end

:remove_vhost
cls
echo --- Menghapus Virtual Host untuk %VHOST_DOMAIN% ---
echo.

:: Menghapus entri dari file hosts
echo [PROSES] Menghapus entri '%VHOST_DOMAIN%' dari file hosts...
(for /f "delims=" %%i in ('findstr /v /c:"127.0.0.1 %VHOST_DOMAIN%" "%SYSTEMROOT%\System32\drivers\etc\hosts"') do @echo %%i) > "%SYSTEMROOT%\System32\drivers\etc\hosts.tmp"
move /y "%SYSTEMROOT%\System32\drivers\etc\hosts.tmp" "%SYSTEMROOT%\System32\drivers\etc\hosts" > nul
echo [OK] Entri hosts telah dihapus.

:: Menghapus Port Forwarding
echo [PROSES] Menghapus port forwarding...
netsh interface portproxy delete v4tov4 listenport=80 listenaddress=127.0.0.1 >nul
echo [OK] Port forwarding telah dihapus.
echo.
echo --- PENGHAPUSAN SELESAI ---
goto :end

:end
echo.
pause