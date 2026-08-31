# Reqly - Windows installer (amd64/arm64, PowerShell)
# Usage: irm https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.ps1 | iex
# Or: powershell -ExecutionPolicy Bypass -File install.ps1
param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\reqly"
)

$ErrorActionPreference = "Stop"
$Repo = "Its-Satyajit/reqly"
$BinName = "reqly.exe"

function Get-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($arch -eq "ARM64") { return "arm64" }
    return "amd64" # Default to amd64 for x86_64 / Intel / AMD
}

function Get-DownloadUrl {
    param([string]$Arch, [string]$Version, [string]$Ext)
    if ($Version -eq "latest") {
        return "https://github.com/$Repo/releases/latest/download/reqly-windows-$Arch$Ext"
    } else {
        return "https://github.com/$Repo/releases/download/$Version/reqly-windows-$Arch$Ext"
    }
}

function Install-Reqly {
    $arch = Get-Arch
    Write-Host "Detected Windows ($arch)..." -ForegroundColor Cyan
    # Ensure TLS 1.2
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

    $tmpDir = Join-Path $env:TEMP "reqly-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    try {
        $exeUrl = Get-DownloadUrl -Arch $arch -Version $Version -Ext ".exe"
        $zipUrl = Get-DownloadUrl -Arch $arch -Version $Version -Ext ".zip"
        $exePath = Join-Path $tmpDir $BinName
        $zipPath = Join-Path $tmpDir "reqly.zip"

        Write-Host "Downloading Reqly CLI ($arch)..." -ForegroundColor Gray
        try {
            Invoke-WebRequest -Uri $exeUrl -OutFile $exePath -UseBasicParsing -ErrorAction Stop
            $bin = Get-Item $exePath
        } catch {
            Write-Host "Direct executable not found, trying zip archive..." -ForegroundColor Yellow
            Invoke-WebRequest -Uri $zipUrl -OutFile $zipPath -UseBasicParsing
            Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force
            $bin = Get-ChildItem -Path $tmpDir -Filter $BinName -Recurse | Select-Object -First 1
            if (-not $bin) { $bin = Get-ChildItem -Path $tmpDir -Filter "reqly*.exe" -Recurse | Where-Object { -not $_.PSIsContainer } | Select-Object -First 1 }
        }

        if (-not $bin -or -not (Test-Path $bin.FullName)) {
            throw "Failed to find binary after download"
        }

        # Install to InstallDir
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        $dest = Join-Path $InstallDir $BinName
        Copy-Item -Path $bin.FullName -Destination $dest -Force
        Unblock-File -Path $dest -ErrorAction SilentlyContinue

        # Add to PATH (user)
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$InstallDir*") {
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
            $env:Path += ";$InstallDir"
            Write-Host "Added $InstallDir to PATH (restart terminal to apply)" -ForegroundColor Green
        }

        Write-Host "Installed $BinName to $dest" -ForegroundColor Green
        & $dest --version
        Write-Host "Run: reqly --version" -ForegroundColor Cyan
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Install-Reqly
