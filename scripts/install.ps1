$ErrorActionPreference = "Stop"
$repo = "dibakshya/tokensense"

$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name -replace '^v', ''
$arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "amd64" }
$url = "https://github.com/$repo/releases/download/v$version/tokensense_${version}_windows_${arch}.zip"

Write-Host "Downloading Tokensense $version for Windows/$arch..."
$tmpDir = Join-Path $env:TEMP "tokensense-install"
New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null
$zipPath = Join-Path $tmpDir "tokensense.zip"

Invoke-WebRequest -Uri $url -OutFile $zipPath
Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force

$installDir = Join-Path $env:LOCALAPPDATA "Tokensense"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Copy-Item (Join-Path $tmpDir "tokensense.exe") -Destination $installDir -Force

# Add to PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added $installDir to PATH"
}

Remove-Item -Recurse -Force $tmpDir
Write-Host "Tokensense $version installed. Run: tokensense setup"
