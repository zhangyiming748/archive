#!/bin/bash
ROOT_DIR="$1"

find "$ROOT_DIR" -type f -iname '*.wmv' | while IFS= read -r src; do
  dst="${src%.*}_tmp.mp4"
  echo "==> 转换: $src"
  ffmpeg -i "$src" \
    -c:v libx265 -tag:v hvc1 \
    -c:a libopus -b:a 160k -application audio \
    -preset slow -crf 24 \
    -pix_fmt yuv420p \
    -x265-params "aq-mode=3:aq-strength=1.2:psy-rd=2.0:psy-rdoq=2.0:rdoq-level=1" \
    -map_chapters -1 \
    "$dst"
done