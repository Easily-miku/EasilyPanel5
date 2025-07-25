@echo off
setlocal enabledelayedexpansion

REM EasilyPanel5 全平台编译脚本 (Windows版本)
REM 支持 Windows, Linux, macOS (x64 和 ARM64)

echo ========================================
echo     EasilyPanel5 全平台编译脚本
echo ========================================
echo.

REM 项目信息
set PROJECT_NAME=easilypanel
set VERSION=1.0.0
set BUILD_DIR=build
set SOURCE_DIR=./cmd

REM 检查Go环境
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go 未安装或不在 PATH 中
    pause
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo [INFO] 检测到 Go 版本: %GO_VERSION%

REM 清理构建目录
if exist "%BUILD_DIR%" (
    echo [INFO] 清理构建目录...
    rmdir /s /q "%BUILD_DIR%"
)
mkdir "%BUILD_DIR%"

echo [INFO] 开始全平台编译...
echo.

REM 构建计数
set SUCCESS_COUNT=0
set TOTAL_COUNT=9

REM Windows x64
echo [INFO] 构建 Windows-x64 (windows/amd64)...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-windows-amd64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%.exe" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo @echo off > "%OUTPUT_DIR%\start.bat"
    echo echo Starting EasilyPanel5... >> "%OUTPUT_DIR%\start.bat"
    echo %PROJECT_NAME%.exe >> "%OUTPUT_DIR%\start.bat"
    echo pause >> "%OUTPUT_DIR%\start.bat"
    echo [SUCCESS] Windows-x64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Windows-x64 构建失败
)
echo.

REM Windows ARM64
echo [INFO] 构建 Windows-ARM64 (windows/arm64)...
set GOOS=windows
set GOARCH=arm64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-windows-arm64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%.exe" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo @echo off > "%OUTPUT_DIR%\start.bat"
    echo echo Starting EasilyPanel5... >> "%OUTPUT_DIR%\start.bat"
    echo %PROJECT_NAME%.exe >> "%OUTPUT_DIR%\start.bat"
    echo pause >> "%OUTPUT_DIR%\start.bat"
    echo [SUCCESS] Windows-ARM64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Windows-ARM64 构建失败
)
echo.

REM Linux x64
echo [INFO] 构建 Linux-x64 (linux/amd64)...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-linux-amd64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] Linux-x64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Linux-x64 构建失败
)
echo.

REM Linux ARM64
echo [INFO] 构建 Linux-ARM64 (linux/arm64)...
set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-linux-arm64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] Linux-ARM64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Linux-ARM64 构建失败
)
echo.

REM Linux ARM
echo [INFO] 构建 Linux-ARM (linux/arm)...
set GOOS=linux
set GOARCH=arm
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-linux-arm
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] Linux-ARM 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] Linux-ARM 构建失败
)
echo.

REM macOS x64
echo [INFO] 构建 macOS-x64 (darwin/amd64)...
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-darwin-amd64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] macOS-x64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] macOS-x64 构建失败
)
echo.

REM macOS ARM64
echo [INFO] 构建 macOS-ARM64 (darwin/arm64)...
set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-darwin-arm64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] macOS-ARM64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] macOS-ARM64 构建失败
)
echo.

REM FreeBSD x64
echo [INFO] 构建 FreeBSD-x64 (freebsd/amd64)...
set GOOS=freebsd
set GOARCH=amd64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-freebsd-amd64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] FreeBSD-x64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] FreeBSD-x64 构建失败
)
echo.

REM OpenBSD x64
echo [INFO] 构建 OpenBSD-x64 (openbsd/amd64)...
set GOOS=openbsd
set GOARCH=amd64
set CGO_ENABLED=0
set OUTPUT_DIR=%BUILD_DIR%\%PROJECT_NAME%-%VERSION%-openbsd-amd64
mkdir "%OUTPUT_DIR%"
go build -ldflags "-s -w" -o "%OUTPUT_DIR%\%PROJECT_NAME%" "%SOURCE_DIR%"
if !errorlevel! equ 0 (
    copy README.md "%OUTPUT_DIR%\" >nul 2>&1
    copy CHANGELOG.md "%OUTPUT_DIR%\" >nul 2>&1
    copy VERSION "%OUTPUT_DIR%\" >nul 2>&1
    echo #!/bin/bash > "%OUTPUT_DIR%\start.sh"
    echo echo "Starting EasilyPanel5..." >> "%OUTPUT_DIR%\start.sh"
    echo ./easilypanel >> "%OUTPUT_DIR%\start.sh"
    echo [SUCCESS] OpenBSD-x64 构建完成
    set /a SUCCESS_COUNT+=1
) else (
    echo [ERROR] OpenBSD-x64 构建失败
)
echo.

REM 显示结果
echo ========================================
echo [SUCCESS] 构建完成: %SUCCESS_COUNT%/%TOTAL_COUNT% 个平台
echo ========================================
echo.
echo [INFO] 构建完成! 输出目录: %BUILD_DIR%
echo.
echo [INFO] 构建的文件:
dir /b "%BUILD_DIR%"
echo.

if %SUCCESS_COUNT% equ %TOTAL_COUNT% (
    echo [SUCCESS] 所有平台构建成功! 🎉
) else (
    echo [WARNING] 部分平台构建失败，请检查错误信息
)

pause
