@echo off
setlocal

:restart
REM kill existing bot.exe
taskkill /im bot.exe /f > nul 2>&1

REM kill anything using port 8000
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8000" ^| findstr "LISTENING"') do (
    taskkill /pid %%a /f > nul 2>&1
)

echo building bot.exe...
go build -o bot.exe
if errorlevel 1 (
    echo build failed!
    pause
    goto restart
)

echo build complete. starting bot.exe...
bot.exe

echo bot.exe exited, restarting...
timeout /t 2 /nobreak > nul
goto restart
