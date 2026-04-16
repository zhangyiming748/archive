# H.265 视频转换指南

**平衡推荐（大多数人最合适）**：

```bash
ffmpeg -i "input.mp4" \
  -c:v libx265 \
  -preset slow \
  -crf 23 \
  -pix_fmt yuv420p10le \
  -x265-params "aq-mode=2:psy-rd=2.0:psy-rdoq=2.0" \
  -c:a copy \
  "output_h265.mkv"
```

> **参数说明**：
>
> - `-preset slow`: 或 slower / veryslow，速度越慢压缩率越高
> - `-crf 23`: 从 23~25 开始测试，调高=更小文件，调低=更好质量
> - `-pix_fmt yuv420p10le`: 10-bit 编码，强烈推荐
> - `-x265-params "aq-mode=2:psy-rd=2.0:psy-rdoq=2.0"`: 自适应量化 + 心理视觉优化
