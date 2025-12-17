@echo off
REM Build script for Windows (alternative to Makefile)

echo Building Ollama Code for Windows...

REM Create build directory
if not exist "build" mkdir build

REM Build the binary
go build -ldflags="-s -w" -trimpath -o build\ollama-code.exe .\cmd\ollama-code

if %errorlevel% equ 0 (
    echo.
    echo ✅ Build complete: build\ollama-code.exe
    echo.
    dir build\ollama-code.exe
) else (
    echo.
    echo ❌ Build failed!
    exit /b 1
)
