# EasilyPanel5

<div align="center">

![EasilyPanel5](https://img.shields.io/badge/EasilyPanel-v1.0.0-blue?style=for-the-badge)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![跨平台](https://img.shields.io/badge/跨平台-Windows%20%7C%20Linux%20%7C%20macOS-lightgrey?style=for-the-badge)](https://github.com/yourusername/EasilyPanel5/releases)

**🎮 跨平台通用游戏服务器管理工具**

*功能强大、易于使用的游戏服务器管理面板*

</div>

## ✨ 主要功能

- 🎮 **游戏服务器管理** - 支持Minecraft Java版/基岩版
- ☕ **Java环境管理** - 自动检测和管理多版本Java
- 🌐 **内网穿透** - 集成OpenFRP，支持命令行启动
- 📦 **服务端下载** - 内置多种下载源
- ⚙️ **配置管理** - 可视化配置编辑
- 📊 **实时监控** - 服务器状态和日志监控

## 📥 下载

| 平台 | 下载链接 |
|------|----------|
| Windows x64 | [easilypanel-windows-amd64.exe](https://github.com/yourusername/EasilyPanel5/releases/latest) |
| Linux x64 | [easilypanel-linux-amd64](https://github.com/yourusername/EasilyPanel5/releases/latest) |
| macOS Intel | [easilypanel-macos-amd64](https://github.com/yourusername/EasilyPanel5/releases/latest) |
| macOS Apple Silicon | [easilypanel-macos-arm64](https://github.com/yourusername/EasilyPanel5/releases/latest) |

## 🚀 快速开始

### Windows
1. 下载 `easilypanel-windows-amd64.exe`
2. 双击运行

### Linux/macOS
```bash
# 下载文件
wget https://github.com/yourusername/EasilyPanel5/releases/latest/download/easilypanel-linux-amd64

# 添加执行权限
chmod +x easilypanel-linux-amd64

# 运行
./easilypanel-linux-amd64
```

## 📖 使用说明

### 1. 创建Minecraft服务器
1. 选择 `实例管理` → `创建实例`
2. 输入服务器名称和端口
3. 选择服务器类型
4. 等待自动下载和配置

### 2. 配置内网穿透
1. 选择 `内网穿透` → `配置OpenFRP`
2. 输入认证令牌和用户访问密钥
3. 创建隧道并启动客户端

### 3. Java环境管理
1. 选择 `Java环境` → `检测Java`
2. 系统自动检测已安装的Java版本
3. 可手动添加自定义Java路径

## ⚙️ 配置文件

主配置文件位于 `./configs/config.yaml`：

```yaml
# 应用配置
app:
  name: "EasilyPanel5"
  data_dir: "./data"
  log_level: "info"

# FRP配置
frp:
  openfrp:
    authorization: "你的认证令牌"
    user_token: "你的用户密钥"

# Java配置
java:
  default_version: "17"
  default_min_memory: "1G"
  default_max_memory: "2G"
```

## 🔧 常见问题

**Q: 程序无法启动？**
A: 确保使用支持交互式界面的终端，如Windows Terminal。

**Q: 无法下载服务端？**
A: 检查网络连接，可能需要科学上网。

**Q: OpenFRP连接失败？**
A: 检查认证令牌和用户密钥是否正确。

**Q: macOS提示安全问题？**
A: 在系统偏好设置中允许运行，或执行：
```bash
sudo xattr -rd com.apple.quarantine easilypanel-macos-amd64
```

## 🛠️ 开发

### 环境要求
- Go 1.24+
- Git

### 本地构建
```bash
git clone https://github.com/yourusername/EasilyPanel5.git
cd EasilyPanel5
go mod tidy
go build -o easilypanel ./cmd
```

## 📁 项目结构

```
EasilyPanel5/
├── cmd/                 # 主程序
├── internal/            # 核心模块
│   ├── config/         # 配置管理
│   ├── instance/       # 实例管理
│   ├── frp/           # 内网穿透
│   ├── java/          # Java环境
│   └── menu/          # 菜单系统
├── configs/            # 配置文件
└── data/              # 数据目录
```

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 发起Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [PromptUI](https://github.com/manifoldco/promptui) - 交互式界面
- [OpenFRP](https://openfrp.net/) - 内网穿透服务
- [Viper](https://github.com/spf13/viper) - 配置管理

## 📞 联系我们

- 🐛 [报告问题](https://github.com/yourusername/EasilyPanel5/issues)
- 💡 [功能建议](https://github.com/yourusername/EasilyPanel5/issues)
- 📖 [项目文档](https://github.com/yourusername/EasilyPanel5/wiki)

---

<div align="center">

**⭐ 觉得有用请给个星标！**

Made with ❤️ by EasilyPanel Team

</div>
