# Reqly - Windows installer (amd64/arm64, PowerShell)
# Copyright (C) 2026 It's Satyajit
# SPDX-License-Identifier: Apache-2.0
# Usage: irm https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.ps1 | iex
#        irm https://raw.githubusercontent.com/Its-Satyajit/reqly/main/install.ps1 | iex; Install-Reqly -Desktop   # desktop app
# Or: powershell -ExecutionPolicy Bypass -File install.ps1 -Desktop
#     powershell -ExecutionPolicy Bypass -File install.ps1 -Version v1.2.0 -Desktop
param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:LOCALAPPDATA\reqly",
    [switch]$Desktop,
    [switch]$App,
    [switch]$Help
)

$ErrorActionPreference = "Stop"
$Repo = "Its-Satyajit/reqly"
$BinName = "reqly.exe"
$DesktopBinName = "Reqly.exe"

if ($Help) {
    Write-Host "Usage: install.ps1 [-Version v1.2.0] [-InstallDir <path>] [-Desktop] [-App]"
    Write-Host "  -Desktop, -App  Install desktop app (installer + binary) instead of CLI only"
    Write-Host "  -Version        Release tag, e.g. v1.2.0 or latest (default: latest)"
    Write-Host "  -InstallDir     Install directory (default: %LOCALAPPDATA%\reqly for CLI, %LOCALAPPDATA%\Programs\Reqly for desktop)"
    exit 0
}
if ($App) { $Desktop = $true }

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

function Get-DesktopUrl {
    param([string]$File, [string]$Version)
    if ($Version -eq "latest") {
        return "https://github.com/$Repo/releases/latest/download/$File"
    } else {
        return "https://github.com/$Repo/releases/download/$Version/$File"
    }
}

function Try-Download {
    param([string]$Url, [string]$OutFile)
    try {
        Write-Host "  Trying $Url ..." -ForegroundColor DarkGray
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing -ErrorAction Stop
        if (Test-Path $OutFile) { return $true }
    } catch {
        Write-Host "    Failed: $($_.Exception.Message)" -ForegroundColor DarkGray
    }
    return $false
}

function Cleanup-OldCli {
    Write-Host "Cleaning up old CLI binaries (both paths)..." -ForegroundColor Gray
    $oldPaths = @(
        "$env:LOCALAPPDATA\reqly\reqly.exe",
        "$env:USERPROFILE\.local\bin\reqly.exe",
        "$env:LOCALAPPDATA\Programs\Reqly\reqly.exe",
        "$InstallDir\reqly.exe"
    )
    foreach ($p in $oldPaths) {
        if (Test-Path $p) {
            try { Remove-Item -Path $p -Force -ErrorAction SilentlyContinue; Write-Host "  removed $p" -ForegroundColor DarkGray } catch {}
        }
    }
}

function Cleanup-OldDesktop {
    Write-Host "Cleaning up old desktop binaries..." -ForegroundColor Gray
    $oldPaths = @(
        "$env:LOCALAPPDATA\Programs\Reqly\Reqly.exe",
        "$env:LOCALAPPDATA\reqly\Reqly.exe",
        "$env:ProgramFiles\Reqly\Reqly.exe",
        "$env:ProgramFiles(x86)\Reqly\Reqly.exe"
    )
    foreach ($p in $oldPaths) {
        if (Test-Path $p) {
            try { Remove-Item -Path $p -Force -ErrorAction SilentlyContinue; Write-Host "  removed $p" -ForegroundColor DarkGray } catch {}
        }
    }
}

function Install-Desktop {
    Cleanup-OldDesktop
    Cleanup-OldCli
    $arch = Get-Arch
    Write-Host "Detected Windows ($arch) - Installing Reqly Desktop..." -ForegroundColor Cyan
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

    $tmpDir = Join-Path $env:TEMP "reqly-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
    try {
        # 1) Try NSIS installer first (best UX, handles WebView2, shortcuts)
        $installerUrls = @(
            (Get-DesktopUrl -File "reqly-windows-$arch-installer.exe" -Version $Version),
            (Get-DesktopUrl -File "Reqly-windows-$arch-installer.exe" -Version $Version),
            (Get-DesktopUrl -File "reqly-$arch-installer.exe" -Version $Version),
            (Get-DesktopUrl -File "Reqly.exe" -Version $Version) # fallback to binary name
        )
        $installerPath = Join-Path $tmpDir "Reqly-installer.exe"
        $downloaded = $false
        foreach ($url in $installerUrls) {
            if (Try-Download -Url $url -OutFile $installerPath) {
                Write-Host "Downloaded installer: $url" -ForegroundColor Green
                $downloaded = $true
                break
            }
        }

        if ($downloaded) {
            # Check if it's actually an installer (size heuristic) vs raw binary
            $isInstaller = (Select-String -Path $installerPath -Pattern "Nullsoft" -Quiet -ErrorAction SilentlyContinue) -or ((Get-Item $installerPath).Length -gt 5MB)
            # Try to run installer silently; if it's just the binary, fallback to binary install
            if ($installerPath -like "*installer.exe" -or $isInstaller) {
                Write-Host "Running installer silently..." -ForegroundColor Gray
                try {
                    # NSIS supports /S for silent; if not installer, this will just fail and we fallback
                    $proc = Start-Process -FilePath $installerPath -ArgumentList "/S" -Wait -PassThru -ErrorAction Stop
                    Write-Host "Desktop installed via installer (exit $($proc.ExitCode))" -ForegroundColor Green
                    # Find installed exe
                    $progFiles = @("$env:LOCALAPPDATA\Programs\Reqly\Reqly.exe", "$env:ProgramFiles\Reqly\Reqly.exe", "$env:ProgramFiles(x86)\Reqly\Reqly.exe", "C:\Program Files\Reqly\Reqly.exe")
                    foreach ($p in $progFiles) {
                        if (Test-Path $p) {
                            Write-Host "Found installed: $p" -ForegroundColor Green
                            return
                        }
                    }
                    # If installer didn't place file where expected, fall through to binary copy
                } catch {
                    Write-Host "Installer run failed, falling back to binary copy: $($_.Exception.Message)" -ForegroundColor Yellow
                }
            }
        }

        # 2) Fallback: desktop binary (Wails, WebView2 required at runtime)
        Write-Host "Downloading Reqly Desktop binary ($arch)..." -ForegroundColor Gray
        $desktopUrls = @(
            (Get-DesktopUrl -File "reqly-desktop-windows-$arch.exe" -Version $Version),
            (Get-DesktopUrl -File "Reqly.exe" -Version $Version),
            (Get-DesktopUrl -File "reqly.exe" -Version $Version),
            (Get-DownloadUrl -Arch $arch -Version $Version -Ext ".exe") # last resort CLI
        )
        $binPath = Join-Path $tmpDir $DesktopBinName
        $binDownloaded = $false
        foreach ($url in $desktopUrls) {
            if (Try-Download -Url $url -OutFile $binPath) {
                Write-Host "Downloaded desktop binary: $url" -ForegroundColor Green
                $binDownloaded = $true
                break
            }
        }

        if (-not $binDownloaded) {
            throw "Failed to download desktop binary after trying multiple URLs"
        }

        # Install binary to Program Files or LOCALAPPDATA
        $desktopInstallDir = "$env:LOCALAPPDATA\Programs\Reqly"
        if (-not (Test-Path $desktopInstallDir)) {
            New-Item -ItemType Directory -Path $desktopInstallDir -Force | Out-Null
        }
        $dest = Join-Path $desktopInstallDir $DesktopBinName
        Copy-Item -Path $binPath -Destination $dest -Force
        Unblock-File -Path $dest -ErrorAction SilentlyContinue

        # Create shortcuts
        try {
            $wsh = New-Object -ComObject WScript.Shell
            $desktopLink = Join-Path ([Environment]::GetFolderPath("Desktop")) "Reqly.lnk"
            $startMenuDir = Join-Path ([Environment]::GetFolderPath("Programs")) "Reqly"
            if (-not (Test-Path $startMenuDir)) { New-Item -ItemType Directory -Path $startMenuDir -Force | Out-Null }
            $startLink = Join-Path $startMenuDir "Reqly.lnk"
            foreach ($link in @($desktopLink, $startLink)) {
                $sc = $wsh.CreateShortcut($link)
                $sc.TargetPath = $dest
                $sc.WorkingDirectory = $desktopInstallDir
                $sc.Description = "Reqly - API development environment"
                $sc.Save()
            }
            Write-Host "Created shortcuts" -ForegroundColor Green
        } catch {
            Write-Host "Shortcut creation failed: $($_.Exception.Message)" -ForegroundColor Yellow
        }

        # Add to PATH
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$desktopInstallDir*") {
            [Environment]::SetEnvironmentVariable("Path", "$userPath;$desktopInstallDir", "User")
            $env:Path += ";$desktopInstallDir"
            Write-Host "Added $desktopInstallDir to PATH (restart terminal)" -ForegroundColor Green
        }

        Write-Host "Installed Reqly Desktop to $dest" -ForegroundColor Green
        Write-Host "Run: `"$dest`" or via Start Menu/Desktop shortcut" -ForegroundColor Cyan
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Install-Reqly {
    if ($Desktop) {
        Install-Desktop
        return
    }

    Cleanup-OldCli
    $arch = Get-Arch
    Write-Host "Detected Windows ($arch) - Installing Reqly CLI..." -ForegroundColor Cyan
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
        $bin = $null
        # Try direct exe with multiple fallbacks (desktop vs cli naming)
        $urls = @(
            $exeUrl,
            (Get-DesktopUrl -File "reqly-windows-$arch.exe" -Version $Version),
            (Get-DesktopUrl -File "reqly.exe" -Version $Version)
        )
        foreach ($url in $urls) {
            if (Try-Download -Url $url -OutFile $exePath) {
                $bin = Get-Item $exePath
                break
            }
        }

        if (-not $bin) {
            Write-Host "Direct executable not found, trying zip archive..." -ForegroundColor Yellow
            $zipUrls = @($zipUrl, (Get-DesktopUrl -File "reqly-windows-$arch.zip" -Version $Version))
            foreach ($zUrl in $zipUrls) {
                if (Try-Download -Url $zUrl -OutFile $zipPath) {
                    Expand-Archive -Path $zipPath -DestinationPath $tmpDir -Force
                    $bin = Get-ChildItem -Path $tmpDir -Filter $BinName -Recurse | Select-Object -First 1
                    if (-not $bin) { $bin = Get-ChildItem -Path $tmpDir -Filter "reqly*.exe" -Recurse | Where-Object { -not $_.PSIsContainer } | Select-Object -First 1 }
                    if ($bin) { break }
                }
            }
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
        Write-Host "Tip: Install desktop app with: irm https://raw.githubusercontent.com/$Repo/main/install.ps1 | iex; Install-Reqly -Desktop" -ForegroundColor DarkGray
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Install-Reqly
