#!/usr/bin/env bash
set -e

if [[ "$OSTYPE" == "linux-gnu"* || "$OSTYPE" == "darwin"* ]]; then
  MODELS_PATH="$HOME/.ollama/models"
  RUNTIME_PATH="$HOME/.ollama/runtime"
elif [[ "$OSTYPE" == "cygwin" || "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
  echo "Please run start.ps1 from PowerShell on Windows."
  exit 1
else
  echo "Unsupported OS: $OSTYPE"
  exit 1
fi

echo "Using Ollama models path: $MODELS_PATH"
echo "Using Ollama runtime path: $RUNTIME_PATH"

# Create a temporary override compose file
TMP_OVERRIDE=$(mktemp /tmp/tinychat.override.XXXXXX.yml)

cat > "$TMP_OVERRIDE" <<EOF
services:
  backend:
    volumes:
      - $MODELS_PATH:/root/.ollama/models:ro
      - $RUNTIME_PATH:/root/.ollama/runtime:ro
EOF

# Run docker compose with both files
docker compose -f docker-compose.yml -f "$TMP_OVERRIDE" up -d --build

# Clean up
rm "$TMP_OVERRIDE"
