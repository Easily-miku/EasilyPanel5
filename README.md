# EasilyPanel5

<div align="center">

![EasilyPanel5 Logo](https://img.shields.io/badge/EasilyPanel-v1.0.0-blue?style=for-the-badge)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=for-the-badge)](https://github.com/yourusername/EasilyPanel5/releases)

**🎮 跨平台通用游戏服务器管理工具**

*一个功能强大、易于使用的游戏服务器管理面板，支持 Minecraft、内网穿透、Java环境管理等功能*

[📥 下载](#-下载) • [🚀 快速开始](#-快速开始) • [📖 文档](#-功能特性) • [🤝 贡献](#-贡献)

</div>

---

## 🌟 功能特性

### 🎯 核心功能
- **🎮 游戏服务器管理**: 支持 Minecraft Java版/基岩版服务器的创建、启动、停止、监控
- **☕ Java环境管理**: 自动检测和管理多版本Java环境
- **🌐 内网穿透**: 集成OpenFRP服务，支持命令行和配置文件两种启动方式
- **📦 服务端下载**: 内置多种游戏服务端下载源，支持Paper、Spigot等
- **⚙️ 配置管理**: 可视化配置编辑，支持配置备份和恢复
- **📊 实时监控**: 服务器状态监控、日志查看、性能统计

### 🔧 技术特性
- **跨平台支持**: Windows、Linux、macOS、ArchLinux
- **现代化UI**: 基于PromptUI的交互式命令行界面
- **零依赖部署**: 单文件可执行程序，无需额外安装
- **配置持久化**: YAML格式配置文件，易于编辑和备份
- **日志系统**: 完整的日志记录和轮转机制

## 📥 下载

### 最新版本 v1.0.0

| 平台 | 架构 | 下载链接 | 大小 |
|------|------|----------|------|
| 🐧 **Linux** | x64 | [easilypanel-linux-amd64](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-linux-amd64) | ~9.6MB |
| 🐧 **Linux** | ARM64 | [easilypanel-linux-arm64](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-linux-arm64) | ~9.1MB |
| 🏛️ **ArchLinux** | x64 | [easilypanel-archlinux-amd64](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-archlinux-amd64) | ~9.6MB |
| 🪟 **Windows** | x64 | [easilypanel-windows-amd64.exe](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-windows-amd64.exe) | ~10MB |
| 🪟 **Windows** | ARM64 | [easilypanel-windows-arm64.exe](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-windows-arm64.exe) | ~9.2MB |
| 🍎 **macOS** | Intel | [easilypanel-macos-amd64](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-macos-amd64) | ~9.8MB |
| 🍎 **macOS** | Apple Silicon | [easilypanel-macos-arm64](https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-macos-arm64) | ~9.3MB |

## 🚀 快速开始

### 1. 下载并运行

**Linux/macOS:**
```bash
# 下载对应平台的可执行文件
wget https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-linux-amd64

# 添加执行权限
chmod +x easilypanel-linux-amd64

# 运行
./easilypanel-linux-amd64
```

**Windows:**
```cmd
# 下载 easilypanel-windows-amd64.exe
# 双击运行或在命令行中执行
easilypanel-windows-amd64.exe
```

### 2. 首次配置

程序启动后会自动创建配置目录和必要文件：
```
./data/
├── configs/          # 配置文件
├── instances/        # 服务器实例
├── downloads/        # 下载文件
├── logs/            # 日志文件
└── bin/             # 二进制文件
```

### 3. 基本使用

1. **创建Minecraft服务器**:
   - 选择 `实例管理` → `创建实例`
   - 输入服务器名称和端口
   - 选择服务器类型（Java版/基岩版）
   - 系统会自动下载并配置服务器

2. **配置内网穿透**:
   - 选择 `内网穿透` → `配置OpenFRP`
   - 输入OpenFRP认证令牌和用户访问密钥
   - 创建隧道并启动frpc客户端

3. **管理Java环境**:
   - 选择 `Java环境` → `检测Java`
   - 系统会自动检测已安装的Java版本
   - 可手动添加自定义Java路径

## 📖 详细文档

### 🎮 游戏服务器管理

#### 支持的服务器类型
- **Minecraft Java版**: Paper, Spigot, Vanilla, Fabric, Forge
- **Minecraft 基岩版**: Bedrock Dedicated Server

#### 实例管理功能
- ✅ 创建/删除实例
- ✅ 启动/停止/重启服务器
- ✅ 实时日志查看
- ✅ 配置文件编辑
- ✅ 性能监控
- ✅ 自动备份

### 🌐 内网穿透

#### OpenFRP集成
- **配置文件方式**: 传统的frpc配置文件启动
- **命令行方式**: 使用 `frpc -u 用户密钥 -p 隧道ID` 快速启动
- **隧道管理**: 创建、编辑、删除隧道
- **节点选择**: 自动获取可用节点列表

#### 支持的协议
- TCP/UDP 端口映射
- HTTP/HTTPS 域名绑定
- 自定义域名和SSL证书

### ☕ Java环境管理

#### 自动检测
- 系统PATH中的Java
- 常见安装路径扫描
- 版本信息识别

#### 手动管理
- 添加自定义Java路径
- 设置默认Java版本
- 内存参数配置

## 🛠️ 开发

### 环境要求
- Go 1.24+
- Git

### 本地构建
```bash
# 克隆仓库
git clone https://github.com/yourusername/EasilyPanel5.git
cd EasilyPanel5

# 安装依赖
go mod tidy

# 构建
go build -o easilypanel ./cmd

# 运行
./easilypanel
```

### 跨平台编译
```bash
# Linux x64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o easilypanel-linux-amd64 ./cmd

# Windows x64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o easilypanel-windows-amd64.exe ./cmd

# macOS x64
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o easilypanel-macos-amd64 ./cmd
```

## 📁 项目结构

```
EasilyPanel5/
├── cmd/                 # 主程序入口
├── internal/            # 内部包
│   ├── config/         # 配置管理
│   ├── instance/       # 实例管理
│   ├── frp/           # 内网穿透
│   ├── java/          # Java环境
│   ├── download/      # 下载管理
│   ├── menu/          # 菜单系统
│   └── logger/        # 日志系统
├── configs/            # 配置文件
├── docs/              # 文档
└── scripts/           # 脚本文件
```

## 🤝 贡献

我们欢迎所有形式的贡献！

### 如何贡献
1. Fork 这个仓库
2. 创建你的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交你的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

### 贡献指南
- 遵循现有的代码风格
- 添加适当的测试
- 更新相关文档
- 确保所有测试通过

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [PromptUI](https://github.com/manifoldco/promptui) - 交互式命令行界面
- [Viper](https://github.com/spf13/viper) - 配置管理
- [Logrus](https://github.com/sirupsen/logrus) - 日志系统
- [OpenFRP](https://openfrp.net/) - 内网穿透服务
- [FastMirror](https://fastmirror.net/) - 我的世界服务器核心镜像站
- 
## 🔧 配置说明

### 配置文件位置
- **主配置**: `./configs/config.yaml`
- **实例配置**: `./data/instances/`
- **Java配置**: `./data/configs/detected_java.json`

### 主要配置项
```yaml
# 应用配置
app:
  name: "EasilyPanel5"
  version: "1.0.0"
  data_dir: "./data"
  log_level: "info"

# FRP配置
frp:
  openfrp:
    api_url: "https://api.openfrp.net"
    authorization: "your_auth_token"
    user_token: "your_user_token"
    server_addr: "frp.openfrp.net"
    default_node_id: 1

# Java配置
java:
  default_version: "17"
  default_min_memory: "1G"
  default_max_memory: "2G"
```

## 🚨 常见问题

### Q: 程序启动后没有反应？
A: 请检查终端是否支持交互式界面，建议使用现代终端如Windows Terminal、iTerm2等。

### Q: 无法下载服务端文件？
A: 请检查网络连接，某些下载源可能需要科学上网。可以尝试切换下载源。

### Q: OpenFRP连接失败？
A: 请确认：
1. 认证令牌和用户访问密钥是否正确
2. 网络是否正常
3. OpenFRP服务是否可用

### Q: Java检测不到？
A: 请确认：
1. Java是否正确安装
2. JAVA_HOME环境变量是否设置
3. 可以手动添加Java路径

### Q: macOS提示"无法验证开发者"？
A: 在系统偏好设置 → 安全性与隐私中允许运行，或使用命令：
```bash
sudo xattr -rd com.apple.quarantine easilypanel-macos-amd64
```

## 📊 系统要求

### 最低要求
- **内存**: 512MB RAM
- **存储**: 100MB 可用空间
- **网络**: 互联网连接（用于下载和内网穿透）

### 推荐配置
- **内存**: 2GB+ RAM
- **存储**: 1GB+ 可用空间
- **CPU**: 双核处理器

### 支持的操作系统
- Windows 10/11 (x64/ARM64)
- Linux (各主流发行版)
- macOS 10.15+ (Intel/Apple Silicon)
- ArchLinux

## 🔄 更新日志

### v1.0.0 (2024-07-26)
- 🎉 首次发布
- ✨ 支持Minecraft服务器管理
- ✨ 集成OpenFRP内网穿透
- ✨ Java环境自动检测
- ✨ 现代化交互式界面
- ✨ 跨平台支持

## 🛣️ 开发路线图

### v1.1.0 (计划中)
- [ ] Web管理界面
- [ ] 插件管理系统
- [ ] 更多游戏服务器支持
- [ ] 数据库集成
- [ ] API接口

### v1.2.0 (计划中)
- [ ] 集群管理
- [ ] 监控告警
- [ ] 自动化部署
- [ ] 容器化支持

## 📞 支持

- 🐛 [报告Bug](https://github.com/Easily-miku/EasilyPanel5/issues)
- 💡 [功能建议](https://github.com/Easily-miku/EasilyPanel5/issues)
- 📖 [Wiki文档](https://github.com/Easily-miku/EasilyPanel5/wiki)
- 💬 [讨论区](https://github.com/Easily-miku/EasilyPanel5/discussions)

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Easily-miku/EasilyPanel5&type=Date)](https://star-history.com/#Easily-miku/EasilyPanel5&Date)

---

<div align="center">

**⭐ 如果这个项目对你有帮助，请给我们一个星标！**

Made with ❤️ by EasilyPanel Team - 轻易

</div>
