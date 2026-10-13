#!/bin/bash
ROOT_DIR="$1"

find "$ROOT_DIR" -type f -iname '*.avi' | while IFS= read -r src; do
  dst="${src%.*}_tmp.mp4"
  echo "==> 转换: $src"
  common_args=(
    -c:v libx265 -tag:v hvc1
    -c:a libopus -b:a 160k -application audio
    -af 'pan=stereo|c0=FL+0.5*FC+0.5*BL+0.5*SL|c1=FR+0.5*FC+0.5*BR+0.5*SR'
    -preset slow -crf 24
    -pix_fmt yuv420p
    -x265-params "aq-mode=3:aq-strength=1.2:psy-rd=2.0:psy-rdoq=2.0:rdoq-level=1"
    -map_chapters -1
  )
  # 如果视频带字幕流,字幕轨必须原样(copy)保留,不转码也不丢弃
  if ffprobe -v error -select_streams s -show_entries stream=index -of csv=p=0 "$src" | grep -q .; then
    echo "==> 检测到字幕流,全部字幕轨原样(copy)保留: $src"
    ffmpeg -i "$src" \
      "${common_args[@]}" \
      -map '0:v?' -map '0:a?' -map '0:s?' \
      -c:s copy \
      "$dst"
  else
    ffmpeg -i "$src" \
      "${common_args[@]}" \
      "$dst"
  fi
done
