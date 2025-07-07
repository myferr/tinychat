# start.ps1
# Run this in PowerShell (not cmd.exe)

# Get the current user profile folder
$UserProfile = $env:USERPROFILE

# Define Ollama model and runtime paths on Windows (adjust if your paths differ)
$ModelsPath = Join-Path $UserProfile ".ollama\models"
$RuntimePath = Join-Path $UserProfile ".ollama\runtime"

Write-Host "Using Ollama models path: $ModelsPath"
Write-Host "Using Ollama runtime path: $RuntimePath"

# Create a temporary override file
$TempOverrideFile = New-TemporaryFile

$yaml = @"
services:
  backend:
    volumes:
      - $ModelsPath:/root/.ollama/models:ro
      - $RuntimePath:/root/.ollama/runtime:ro
"@

# Write YAML content to temp file
Set-Content -Path $TempOverrideFile -Value $yaml -Encoding UTF8

# Run docker compose with the override file
docker compose -f docker-compose.yml -f $TempOverrideFile.FullName up -d --build

# Cleanup temp file
Remove-Item $TempOverrideFile
