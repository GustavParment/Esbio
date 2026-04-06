@echo off
:: Esbio Startup Script for Windows
:: Starts the database, backend server, and frontend client

echo.
echo ========================================
echo    Starting Esbio...
echo ========================================
echo.

:: Get the script directory
set SCRIPT_DIR=%~dp0

:: Check if PostgreSQL container is running
echo [1/3] Checking PostgreSQL container...
docker ps | findstr "esbio-postgres" >nul 2>&1
if %errorlevel%==0 (
    echo       PostgreSQL container is already running
) else (
    echo       Starting PostgreSQL container...
    cd /d "%SCRIPT_DIR%server"
    docker-compose up -d
    echo       Waiting for PostgreSQL to be ready...
    timeout /t 3 /nobreak >nul
)
echo.

:: Start the backend server in a new window
echo [2/3] Starting Go backend server on :8080...
cd /d "%SCRIPT_DIR%server\cmd"
start "Esbio Backend" cmd /k "go run api/main.go"
echo       Backend starting in new window...
timeout /t 2 /nobreak >nul
echo.

:: Start the frontend client in a new window
echo [3/3] Starting Next.js frontend on :3000...
cd /d "%SCRIPT_DIR%client"
start "Esbio Frontend" cmd /k "npm run dev"
echo       Frontend starting in new window...
echo.

echo ========================================
echo    Esbio is now running!
echo.
echo    Frontend:  http://localhost:3000
echo    Backend:   http://localhost:8080
echo    Database:  postgres://localhost:5433
echo.
echo    Each service runs in its own window.
echo    Close those windows to stop services.
echo ========================================
echo.
pause
