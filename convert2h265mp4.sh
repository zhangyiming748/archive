#!/bin/bash
ROOT_DIR="$1"

find "$ROOT_DIR" -type f -iname '*.avi' | while IFS= read -r src; do
  dst="${src%.*}_tmp.mp4"
  echo "==> 转换: $src"
  ffmpeg -i "$src" \
    -c:v libx265 -tag:v hvc1 \
    -c:a libopus -b:a 160k -application audio \
    -af 'pan=stereo|c0=FL+0.5*FC+0.5*BL+0.5*SL|c1=FR+0.5*FC+0.5*BR+0.5*SR' \
    -preset slow -crf 24 \
    -pix_fmt yuv420p \
    -x265-params "aq-mode=3:aq-strength=1.2:psy-rd=2.0:psy-rdoq=2.0:rdoq-level=1" \
    -map_chapters -1 \
    "$dst"
done