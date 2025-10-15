# Archive

一个用于媒体文件整理和转码的Go工具库。

## 功能

本项目提供以下主要功能：

1. 视频文件转码为H.265/HEVC格式（带hvc1标签的MP4文件）
2. 图片文件转码为AVIF格式

## 函数说明

### Convert2H265

将视频文件转换为H.265/HEVC编码的MP4文件，并添加hvc1标签。

特性：
- 自动检测已经是H.265编码的视频并避免重复转换
- 对于已经是H.265但缺少hvc1标签的视频，仅添加标签而不重新编码
- 超高清视频（超过1920像素）自动缩放到合适尺寸
- 显示转换前后的文件大小差异
- 转换完成后自动删除源文件

### Convert2AVIF

将图片文件转换为AVIF格式，这是一种现代高效的图像格式。

特性：
- 使用libaom-av1编码器进行转换
- 转换完成后显示文件大小差异
- 自动删除源文件

## 依赖

- ffmpeg：用于音视频编解码
- github.com/zhangyiming748/FastMediaInfo：用于获取媒体文件信息

## 安装

```bash
go get github.com/zhangyiming748/archive
```

## 使用示例

```go
package main

import "github.com/zhangyiming748/archive"

func main() {
    // 转换视频文件
    archive.Convert2H265("/path/to/video.mp4")
    
    // 转换图片文件
    archive.Convert2AVIF("/path/to/image.jpg")
}
```

## 许可证

本项目采用MIT许可证。详情请见[LICENSE](LICENSE)文件。
