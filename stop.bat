@echo off
:: Esbio Stop Script for Windows
:: Stops the backend server, frontend client, and optionally the database

echo.
echo ========================================
echo    Stopping Esbio...
echo ========================================
echo.

:: Kill Go backend processes
echo [1/3] Stopping Go backend...
taskkill /f /im go.exe >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq Esbio Backend*" >nul 2>&1
echo       Backend stopped
echo.

:: Kill Node.js frontend processes
echo [2/3] Stopping Next.js frontend...
taskkill /f /im node.exe >nul 2>&1
taskkill /f /fi "WINDOWTITLE eq Esbio Frontend*" >nul 2>&1
echo       Frontend stopped
echo.

:: Ask about database
echo [3/3] PostgreSQL container...
set /p STOP_DB="       Stop database container too? (y/N): "
if /i "%STOP_DB%"=="y" (
    docker stop esbio-postgres >nul 2>&1
    echo       Database stopped
) else (
    echo       Database still running
)
echo.

echo ========================================
echo    Esbio has been stopped
echo ========================================
echo.
pause
