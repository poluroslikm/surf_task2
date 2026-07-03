#!/usr/bin/env bash
# Сборка PDF дизайн-брифа «Волна»: md -> HTML (Python) -> PDF (Chrome headless).
set -euo pipefail

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HTML="$DIR/design-brief.html"
PDF="$DIR/design-brief.pdf"
CHROME="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

echo "==> 1/2  Сборка HTML"
python3 "$DIR/build.py"

echo "==> 2/2  Печать в PDF (Chrome headless)"
"$CHROME" --headless=new --disable-gpu --no-pdf-header-footer \
  --print-to-pdf="$PDF" "file://$HTML" 2>/dev/null

echo "Готово: $PDF"
ls -lh "$PDF" | awk '{print "Размер:", $5}'
