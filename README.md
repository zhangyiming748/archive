# Archive

一个用于媒体文件整理和转码的Go工具库。


## 函数说明
+ ConvertMKV2H265 转换MKV文件为libx265的H.265/HEVC格式 音频为aac 保留全部字幕轨
+ Convert2H265 将视频文件转换为H.265/HEVC编码的MP4文件，并添加hvc1标签。
+ Convert2AVIF 将图片文件转换为AVIF格式，这是一种现代高效的图像格式。
+ FastConvertVideo2StandAvc 快速转换视频文件为标准H264(avc)视频

### 使用示例

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

