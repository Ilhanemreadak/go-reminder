@echo off
echo [*] Email Reminder System kapatiliyor...

:: Kill reminder.exe process
tasklist /FI "IMAGENAME eq reminder.exe" 2>NUL | find /I "reminder.exe" >NUL
if %ERRORLEVEL%==0 (
    taskkill /F /IM reminder.exe >NUL 2>&1
    echo [OK] Uygulama kapatildi.
) else (
    echo [*] reminder.exe bulunamadi.
)

:: Also kill any process using port 9147
for /f "tokens=5" %%a in ('netstat -ano ^| findstr "LISTENING" ^| findstr ":9147"') do (
    echo [*] Port 9147 kullanan process kapatiliyor (PID: %%a)...
    taskkill /F /PID %%a >NUL 2>&1
    echo [OK] Process kapatildi.
)

echo [*] Tamamlandi.
timeout /t 3 /nobreak >NUL
