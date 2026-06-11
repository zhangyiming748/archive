# Archive - 媒体文件智能归档与转码工具

[![Go Version](https://img.shields.io/github/go-mod/go-version/zhangyiming748/archive)](https://golang.org/)
[![License](https://img.shields.io/github/license/zhangyiming748/archive)](LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/zhangyiming748/archive)](https://github.com/zhangyiming748/archive/stargazers)

Archive 是一个基于 Go 语言开发的专业级媒体文件整理与转码工具库，专注于帮助用户高效地将视频、音频和图像文件转换为更现代、更节省空间的编码格式。

## 🎯 核心功能

### 视频处理
- **MKV 转 H.265**: 智能转换 MKV 文件为 H.265/HEVC 格式，完整保留所有音轨和字幕轨道
- **通用视频转码**: 支持任意视频格式转换为 H.265 编码的 MP4 文件，并添加 hvc1 兼容标签
- **快速标准化**: 提供快速 H.264(AVC) 标准化转换，适用于流媒体播放
- **智能格式检测**: 自动识别视频编码格式，避免重复转换

### 图像处理
- **AVIF 转换**: 将各种图像格式高效转换为现代 AVIF 格式，显著减少文件体积
- **智能跳过**: 自动跳过 GIF 等特殊格式文件

### 音频处理
- **有声书优化**: 支持有声书音频加速处理（65%速度）并提升音量
- **音乐处理**: 针对不同类型音频提供定制化处理方案
- **格式标准化**: 统一音频采样率、比特率等参数

### 数据管理
- **SQLite 记录**: 自动记录每次转换的文件信息和空间节省情况
- **历史追踪**: 完整保存转换历史，便于统计和回溯

## 🚀 快速开始

### 系统要求
- Go 1.21 或更高版本
- FFmpeg (必需) - 视频/音频处理引擎
- MediaInfo (推荐) - 媒体信息分析
- libavif (推荐) - AVIF 图像编码

### 安装依赖

#### Windows
```bash
# 安装 FFmpeg
choco install ffmpeg

# 安装 MediaInfo
choco install mediainfo

# 安装 libavif
# 从 https://github.com/AOMediaCodec/libavif 下载预编译版本
```

#### macOS
```bash
brew install ffmpeg mediainfo libavif
```

#### Linux (Ubuntu/Debian)
```bash
sudo apt update
sudo apt install ffmpeg mediainfo
# libavif 需要从源码编译或使用 snap 安装
```

### 安装 Archive

```bash
go get github.com/zhangyiming748/archive@latest
```

## 💡 使用示例

### 基础用法

```go
package main

import (
    "log"
    "github.com/zhangyiming748/archive"
    "github.com/zhangyiming748/finder"
)

func main() {
    // 批量转换视频文件为 H.265
    videos := finder.FindAllVideos("/path/to/videos")
    for _, video := range videos {
        archive.Convert2H265(video)
    }
    
    // 转换单个图片为 AVIF
    archive.Convert2AVIF("/path/to/image.jpg")
    
    // 处理有声书音频
    archive.ConvertAudio("/path/to/audiobook.mp3", archive.AudioBookType)
}
```

### 高级功能

```go
// MKV 文件专用处理（保留所有轨道）
archive.ConvertMKV2H265("/path/to/movie.mkv")

// 快速标准化转换
archive.FastConvertVideo2StandAvc("/path/to/video.avi")

// 字幕内嵌处理
err := archive.MergeMp4WithSameNameSrt("video.mp4", "subtitle.srt")
if err != nil {
    log.Fatal(err)
}
```

## 🔧 API 文档

### 视频处理函数

| 函数 | 描述 | 参数 |
|------|------|------|
| `ConvertMKV2H265(src string)` | MKV 转 H.265，保留所有音轨字幕 | 源文件路径 |
| `Convert2H265(src string)` | 通用视频转 H.265 MP4 | 源文件路径 |
| `CloneMkv2H265(src string)` | MKV 文件克隆转码 | 源文件路径 |
| `FastConvertVideo2StandAvc(src string)` | 快速标准化 AVC 转换 | 源文件路径 |
| `FastConvertMkv(src string)` | 快速 MKV 转 MP4 | 源文件路径 |
| `MergeMp4WithSameNameSrt(video, srt string)` | 内嵌字幕到视频 | 视频和字幕文件路径 |

### 图像处理函数

| 函数 | 描述 | 参数 |
|------|------|------|
| `Convert2AVIF(src string)` | 图片转 AVIF 格式 | 源文件路径 |

### 音频处理函数

| 函数 | 描述 | 参数 |
|------|------|------|
| `ConvertAudio(src, mytype string)` | 音频文件转换处理 | 源文件路径, 音频类型 |

### 常量定义

```go
const (
    AudioBookType = "audiobook"  // 有声书类型
    RapMusicType  = "rap"        // 说唱音乐类型
    Volume        = "2.7"        // 音频音量增益
)
```

## 📊 性能特点

- **智能检测**: 自动跳过已优化的文件，避免重复处理
- **空间节省**: H.265 编码相比 H.264 可节省 30-50% 存储空间
- **质量保持**: 在大幅压缩的同时维持优秀的视觉质量
- **批量处理**: 支持大规模文件批处理操作
- **进度追踪**: 实时显示转换进度和空间节省情况

## 🛠️ 技术架构

```
archive/
├── video.go          # 视频处理核心逻辑
├── image.go          # 图像转换功能
├── audio.go          # 音频处理模块
├── stand.go          # 标准化处理和依赖检查
├── sqlite/           # 数据库存储层
│   ├── model.go      # 数据模型定义
│   ├── sqlite.go     # SQLite 连接管理
│   └── unit_test.go  # 数据库单元测试
└── unit_test.go      # 主要功能测试
```

## 🔍 依赖说明

| 依赖 | 用途 | 版本 |
|------|------|------|
| `github.com/zhangyiming748/FastMediaInfo` | 媒体信息获取 | v0.0.7 |
| `github.com/zhangyiming748/finder` | 文件搜索工具 | v0.0.7 |
| `github.com/glebarez/sqlite` | SQLite 驱动 | v1.11.0 |
| `gorm.io/gorm` | ORM 框架 | v1.31.1 |

## 📝 测试

运行单元测试：

```bash
go test -v ./...
```

运行特定测试：

```bash
# 测试 MKV 转换
go test -v -run TestConvertMkv

# 测试视频标准化
go test -v -run TestConvertH264

# 测试数据库操作
go test -v -run TestHistoryModel
```

## ⚠️ 注意事项

1. **备份重要文件**: 转换过程会删除源文件，请确保重要数据已备份
2. **系统资源**: 大批量转换会消耗较多 CPU 和内存资源
3. **磁盘空间**: 转换过程中需要临时存储空间
4. **权限要求**: 确保程序有读写目标目录的权限

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request 来改进项目：

1. Fork 项目仓库
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

感谢以下开源项目的支持：
- [FFmpeg](https://ffmpeg.org/) - 多媒体处理框架
- [libavif](https://github.com/AOMediaCodec/libavif) - AVIF 图像编码库
- [MediaInfo](https://mediaarea.net/en/MediaInfo) - 媒体信息分析工具
- [GORM](https://gorm.io/) - Go 语言 ORM 框架

---

## 📅 开发日志

### 2026-06
- **修复视频缩放参数** - 使用 FFmpeg `-2` 参数确保输出尺寸为偶数，避免编码错误
- **修复视频分辨率检查逻辑** - 优化 overFHD 函数，移除不必要的取模检查
- **添加视频转换进度条显示功能** - 实时展示转换进度和状态
- **更新项目依赖包** - 升级 Go 模块依赖到最新版本

### 2026-05
- **添加强制覆盖功能** - 视频转换函数支持强制覆盖已存在文件
- **添加 MKV 到 MP4 格式转换功能** - 提供快速 MKV 转 MP4 的工具方法
- **优化数据库连接初始化逻辑** - 改进 SQLite 连接的稳定性和性能
- **修复视频旋转功能** - 修正编解码器配置问题，确保旋转功能正常工作
- **优化 AVIF 编码性能** - 提升图像转换效率并修复视频编码器检测
- **修复视频转换文件名拼接错误** - 解决输出文件名生成问题
- **移除视频编码中的 hvc1 标签设置** - 简化编码流程
- **添加硬件加速视频编码支持** - 利用 GPU 加速视频转码过程
- **添加视频格式日志记录** - 增强调试信息输出
- **优化音频转换临时文件生成逻辑** - 改进临时文件管理

### 2026-04
- **添加大疆视频转换功能** - 专门处理 DJI 无人机拍摄的视频文件
- **添加音频格式转换功能** - 支持多种音频格式之间的转换
- **优化视频转码逻辑** - 添加分辨率检查功能，智能判断是否需要转码
- **优化视频转换参数** - 提升压缩效率和视频质量
- **添加视频提取音频功能** - 从视频中分离音频轨道

### 2026-03
- **重构依赖检查功能** - 更改为手动执行数据库初始化，优化初始化流程
- **更新项目依赖包版本** - 升级 Go 版本和 finder 依赖
- **添加 FHD 参数控制** - 支持控制视频是否缩放到 1080p 分辨率
- **添加文件清理逻辑** - 自动清理转换过程中的临时文件
- **修复图像转换命令参数初始化问题** - 确保参数正确传递
- **支持 BMP 格式图片转换为 AVIF** - 扩展支持的输入图像格式
- **添加 AVIF 文件跳过逻辑** - 避免重复转换已是 AVIF 格式的文件

### 2026-02
- **添加旋转视频的方法** - 支持视频顺时针/逆时针旋转
- **内嵌字幕详细输出** - 字幕内嵌成功或失败都有详细的日志输出
- **更新 README 文档** - 完善使用说明和 API 文档

### 2026-01
- **添加 MergeMp4WithSameNameSrt 方法** - 将同名字幕文件内嵌到视频中
- **修复多余参数问题** - 清理不必要的命令行参数
- **批量更改 MKV 为 MP4** - 提供批量转换工具方法
- **图片转换为 AVIF 使用 avifenc 命令** - 改用官方工具实现转换
- **跳过 GIF 文件处理** - 避免转换不支持的图像格式

### 2025-12
- **取消线程数限制** - 允许并行处理更多文件
- **代码重构** - 明确分工，与 finder 库职责分离
- **整理音频处理方法** - 优化音频转换的代码结构
- **修复 bug** - 多个小问题的修复和优化

### 2025-11
- **快速转换视频为标准 H.264(AVC)** - 便于无损分割和流媒体播放
- **MKV 完整逻辑实现** - 单独设置 MKV 文件的处理逻辑
- **非破坏性添加 MKV 逻辑** - 保留所有音轨和字幕轨道
- **转换 MKV 文件为 H.265 格式** - 保留全部音频轨、字幕轨

### 2025-10
- **数据库文件默认保存位置** - 保存到用户家目录下的 sqlite.db
- **计算差值之前同步表结构** - 确保数据库结构一致性
- **DiffSize 函数公开** - 暴露空间节省计算功能
- **同步表结构** - 自动维护数据库表结构
- **使用 SQLite 记录节省的空间** - 追踪每次转换的空间优化效果
- **实现基础方法** - 完成核心的视频和图像处理功能

### 2025-09
- **Initial commit** - 项目初始化，建立基础架构

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/zhangyiming748">zhangyiming748</a>
</p>