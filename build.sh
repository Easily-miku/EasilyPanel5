#!/bin/bash

# EasilyPanel5 全平台编译脚本
# 支持 Windows, Linux, macOS (x64 和 ARM64)

set -e

# 项目信息
PROJECT_NAME="easilypanel"
VERSION="1.0.0"
BUILD_DIR="build"
SOURCE_DIR="./cmd"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Go环境
check_go() {
    if ! command -v go &> /dev/null; then
        print_error "Go 未安装或不在 PATH 中"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    print_info "检测到 Go 版本: $GO_VERSION"
}

# 清理构建目录
clean_build() {
    if [ -d "$BUILD_DIR" ]; then
        print_info "清理构建目录..."
        rm -rf "$BUILD_DIR"
    fi
    mkdir -p "$BUILD_DIR"
}

# 构建函数
build_for_platform() {
    local GOOS=$1
    local GOARCH=$2
    local EXT=$3
    local PLATFORM_NAME=$4
    
    local OUTPUT_NAME="${PROJECT_NAME}"
    if [ "$EXT" != "" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.${EXT}"
    fi
    
    local OUTPUT_DIR="${BUILD_DIR}/${PROJECT_NAME}-${VERSION}-${GOOS}-${GOARCH}"
    local OUTPUT_PATH="${OUTPUT_DIR}/${OUTPUT_NAME}"
    
    print_info "构建 ${PLATFORM_NAME} (${GOOS}/${GOARCH})..."
    
    # 创建输出目录
    mkdir -p "$OUTPUT_DIR"
    
    # 设置环境变量并构建
    export GOOS=$GOOS
    export GOARCH=$GOARCH
    export CGO_ENABLED=0
    
    if go build -ldflags "-s -w -X main.Version=${VERSION}" -o "$OUTPUT_PATH" "$SOURCE_DIR"; then
        # 复制必要文件
        cp README.md "$OUTPUT_DIR/" 2>/dev/null || true
        cp CHANGELOG.md "$OUTPUT_DIR/" 2>/dev/null || true
        cp VERSION "$OUTPUT_DIR/" 2>/dev/null || true
        
        # 创建启动脚本
        if [ "$GOOS" = "windows" ]; then
            cat > "${OUTPUT_DIR}/start.bat" << 'EOF'
@echo off
echo Starting EasilyPanel5...
easilypanel.exe
pause
EOF
        else
            cat > "${OUTPUT_DIR}/start.sh" << 'EOF'
#!/bin/bash
echo "Starting EasilyPanel5..."
./easilypanel
EOF
            chmod +x "${OUTPUT_DIR}/start.sh"
        fi
        
        # 获取文件大小
        local SIZE=$(du -h "$OUTPUT_PATH" | cut -f1)
        print_success "${PLATFORM_NAME} 构建完成 (大小: $SIZE)"
        
        # 创建压缩包
        create_archive "$OUTPUT_DIR" "$GOOS"
        
    else
        print_error "${PLATFORM_NAME} 构建失败"
        return 1
    fi
}

# 创建压缩包
create_archive() {
    local DIR=$1
    local OS=$2
    local ARCHIVE_NAME=$(basename "$DIR")
    
    cd "$BUILD_DIR"
    
    if [ "$OS" = "windows" ]; then
        # Windows 使用 zip
        if command -v zip &> /dev/null; then
            zip -r "${ARCHIVE_NAME}.zip" "$(basename "$DIR")" > /dev/null
            print_info "创建压缩包: ${ARCHIVE_NAME}.zip"
        fi
    else
        # Unix 系统使用 tar.gz
        tar -czf "${ARCHIVE_NAME}.tar.gz" "$(basename "$DIR")"
        print_info "创建压缩包: ${ARCHIVE_NAME}.tar.gz"
    fi
    
    cd ..
}

# 显示构建结果
show_results() {
    print_info "构建完成! 输出目录: $BUILD_DIR"
    echo
    print_info "构建的文件:"
    ls -la "$BUILD_DIR"
    echo
    
    print_info "文件大小统计:"
    find "$BUILD_DIR" -name "$PROJECT_NAME*" -type f | while read file; do
        size=$(du -h "$file" | cut -f1)
        echo "  $(basename "$file"): $size"
    done
}

# 主函数
main() {
    echo "========================================"
    echo "    EasilyPanel5 全平台编译脚本"
    echo "========================================"
    echo
    
    check_go
    clean_build
    
    print_info "开始全平台编译..."
    echo
    
    # 定义构建目标
    # 格式: GOOS GOARCH 扩展名 平台名称
    TARGETS=(
        "windows amd64 exe Windows-x64"
        "windows arm64 exe Windows-ARM64"
        "linux amd64 '' Linux-x64"
        "linux arm64 '' Linux-ARM64"
        "linux arm '' Linux-ARM"
        "darwin amd64 '' macOS-x64"
        "darwin arm64 '' macOS-ARM64"
        "freebsd amd64 '' FreeBSD-x64"
        "openbsd amd64 '' OpenBSD-x64"
    )
    
    # 构建计数
    SUCCESS_COUNT=0
    TOTAL_COUNT=${#TARGETS[@]}
    
    # 执行构建
    for target in "${TARGETS[@]}"; do
        read -r goos goarch ext platform <<< "$target"
        if build_for_platform "$goos" "$goarch" "$ext" "$platform"; then
            ((SUCCESS_COUNT++))
        fi
        echo
    done
    
    # 显示结果
    echo "========================================"
    print_success "构建完成: $SUCCESS_COUNT/$TOTAL_COUNT 个平台"
    echo "========================================"
    
    show_results
    
    if [ $SUCCESS_COUNT -eq $TOTAL_COUNT ]; then
        print_success "所有平台构建成功! 🎉"
    else
        print_warning "部分平台构建失败，请检查错误信息"
    fi
}

# 运行主函数
main "$@"
