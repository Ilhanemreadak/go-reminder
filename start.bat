@echo off
title Email Reminder System

:: Kill any existing instance
tasklist /FI "IMAGENAME eq reminder.exe" 2>NUL | find /I "reminder.exe" >NUL
if %ERRORLEVEL%==0 (
    echo [*] Mevcut uygulama kapatiliyor...
    taskkill /F /IM reminder.exe >NUL 2>&1
    timeout /t 2 /nobreak >NUL
)

:: Also kill any go run main.go processes using the port
for /f "tokens=5" %%a in ('netstat -ano ^| findstr "LISTENING" ^| findstr ":9147"') do (
    echo [*] Port 9147 kullanan process kapatiliyor (PID: %%a)...
    taskkill /F /PID %%a >NUL 2>&1
)

timeout /t 1 /nobreak >NUL

:: Build and start the application
echo [*] Uygulama derleniyor...
cd /d "%~dp0"
go build -o reminder.exe .
if %ERRORLEVEL% NEQ 0 (
    echo [!] Derleme hatasi!
    pause
    exit /b 1
)

echo [*] Uygulama baslatiliyor (port 9147)...
echo [*] Tarayicida acin: http://localhost:9147
echo [*] Kapatmak icin stop.bat dosyasini calistirin veya bu pencereyi kapatin.
echo.
reminder.exe
