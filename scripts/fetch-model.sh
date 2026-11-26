#!/usr/bin/env bash
# Download a small Qwen-class GGUF into ./models/model.gguf for llama.cpp.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
mkdir -p "$ROOT/models"
DEST="$ROOT/models/model.gguf"

# Default: Qwen2.5 0.5B Instruct Q4_K_M (Hugging Face mirror via huggingface-cli or curl).
# Override with MODEL_URL to pin another GGUF.
MODEL_URL="${MODEL_URL:-https://huggingface.co/bartowski/Qwen2.5-0.5B-Instruct-GGUF/resolve/main/Qwen2.5-0.5B-Instruct-Q4_K_M.gguf}"

if [[ -f "$DEST" ]]; then
  echo "already present: $DEST"
  exit 0
fi

echo "fetching $MODEL_URL"
if command -v curl >/dev/null 2>&1; then
  curl -fL --progress-bar -o "$DEST.partial" "$MODEL_URL"
  mv "$DEST.partial" "$DEST"
elif command -v wget >/dev/null 2>&1; then
  wget -O "$DEST.partial" "$MODEL_URL"
  mv "$DEST.partial" "$DEST"
else
  echo "need curl or wget" >&2
  exit 1
fi
echo "saved $DEST"
