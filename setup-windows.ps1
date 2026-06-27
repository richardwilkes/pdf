<#
.SYNOPSIS
    Sets up a fresh Windows machine to build github.com/richardwilkes/pdf.

.DESCRIPTION
    This package uses cgo to link against the vendored static MuPDF libraries in
    lib/. Those Windows libraries are built with a UCRT mingw-w64 toolchain and
    reference symbols such as __intrinsic_setjmpex that are emitted only when
    compiling against UCRT mingw-w64 headers and are resolvable only by a UCRT
    mingw-w64 runtime.

    This script reproduces the CI toolchain locally:
      * Git                (Git.Git)       - matches actions/checkout
      * Go                 (GoLang.Go)     - matches actions/setup-go (go.mod => 1.26+)
      * UCRT mingw-w64 gcc (MSYS2 ucrt64)  - the C compiler/linker cgo needs
    It then puts the ucrt64 compiler on your PATH and sets CGO_ENABLED=1 so an
    ordinary Git Bash (or PowerShell) shell can run ./build.sh.

    Run this from a normal PowerShell prompt. Re-running is safe (idempotent).
    IMPORTANT: build from Git Bash or PowerShell afterward, NOT from the MSYS2
    shell. Git Bash ships its own MSYS2 runtime; it only needs the mingw-w64
    gcc visible on PATH, which this script arranges.

.NOTES
    Open a NEW terminal after this finishes so the updated PATH/env take effect.
#>

[CmdletBinding()]
param(
    # MSYS2 install root. Default matches the MSYS2.MSYS2 winget package.
    [string]$Msys64Root = 'C:\msys64'
)

$ErrorActionPreference = 'Stop'

function Write-Step([string]$msg)  { Write-Host "==> $msg" -ForegroundColor Cyan }
function Write-Ok([string]$msg)    { Write-Host "    $msg"   -ForegroundColor Green }
function Write-Warn2([string]$msg) { Write-Host "    $msg"  -ForegroundColor Yellow }

# --- Preconditions ---------------------------------------------------------

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
    throw "winget is required but was not found. Install 'App Installer' from the Microsoft Store, then re-run."
}

# Install a winget package only if it isn't already present.
function Install-WingetPackage([string]$id, [string]$friendly) {
    Write-Step "Checking $friendly ($id)"
    winget list --id $id --exact --accept-source-agreements *> $null
    if ($LASTEXITCODE -eq 0) {
        Write-Ok "$friendly already installed."
        return
    }
    Write-Step "Installing $friendly"
    winget install --id $id --exact --silent `
        --accept-package-agreements --accept-source-agreements `
        --disable-interactivity
    if ($LASTEXITCODE -ne 0) {
        throw "winget failed to install $friendly ($id). Exit code $LASTEXITCODE."
    }
    Write-Ok "$friendly installed."
}

# --- Git, Go, MSYS2 --------------------------------------------------------

Install-WingetPackage -id 'Git.Git'     -friendly 'Git'
Install-WingetPackage -id 'GoLang.Go'   -friendly 'Go'
Install-WingetPackage -id 'MSYS2.MSYS2' -friendly 'MSYS2'

# --- mingw-w64 gcc via MSYS2's pacman --------------------------------------

$msysBash = Join-Path $Msys64Root 'usr\bin\bash.exe'
if (-not (Test-Path $msysBash)) {
    throw "MSYS2 bash not found at $msysBash. If MSYS2 installed elsewhere, re-run with -Msys64Root <path>."
}

$mingwBin = Join-Path $Msys64Root 'ucrt64\bin'
$gccExe   = Join-Path $mingwBin 'gcc.exe'

if (Test-Path $gccExe) {
    Write-Step "Checking UCRT mingw-w64 gcc"
    Write-Ok "gcc already present at $gccExe."
} else {
    Write-Step "Installing UCRT mingw-w64 gcc through MSYS2 (pacman)"
    # -Sy refreshes the package DB; --needed skips work if already current;
    # --noconfirm keeps it non-interactive. mingw-w64-ucrt-x86_64-gcc is the UCRT
    # toolchain (ucrt64). The MSVCRT variant (mingw-w64-x86_64-gcc / mingw64) does
    # NOT resolve __intrinsic_setjmpex against the vendored MuPDF libs.
    & $msysBash -lc "pacman -Sy --noconfirm && pacman -S --needed --noconfirm mingw-w64-ucrt-x86_64-gcc"
    if ($LASTEXITCODE -ne 0) {
        throw "pacman failed to install mingw-w64-ucrt-x86_64-gcc. Exit code $LASTEXITCODE."
    }
    if (-not (Test-Path $gccExe)) {
        throw "gcc still not found at $gccExe after install."
    }
    Write-Ok "UCRT mingw-w64 gcc installed."
}

# --- Persist environment for future shells ---------------------------------

# Remove directories from the persistent (User) PATH (idempotent cleanup).
function Remove-FromUserPath([string[]]$dirs) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ([string]::IsNullOrEmpty($current)) { return }
    $parts = $current -split ';' | Where-Object { $_ -ne '' -and ($dirs -notcontains $_) }
    [Environment]::SetEnvironmentVariable('Path', ($parts -join ';'), 'User')
}

# Prepend a directory to the persistent (User) PATH if not already present.
function Add-ToUserPath([string]$dir) {
    $current = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ([string]::IsNullOrEmpty($current)) { $current = '' }
    $parts = $current -split ';' | Where-Object { $_ -ne '' }
    if ($parts -contains $dir) {
        Write-Ok "PATH already contains $dir"
        return
    }
    $new = (@($dir) + $parts) -join ';'
    [Environment]::SetEnvironmentVariable('Path', $new, 'User')
    Write-Ok "Added to user PATH: $dir"
}

Write-Step "Configuring PATH and CGO_ENABLED for future shells"

# Drop the MSVCRT mingw64 entry an earlier version of this script may have added;
# it links the wrong C runtime for the vendored MuPDF libs.
Remove-FromUserPath @((Join-Path $Msys64Root 'mingw64\bin'))

Add-ToUserPath $mingwBin

# Go's per-user tool bin (where build.sh installs golangci-lint).
$goBin = Join-Path $env:USERPROFILE 'go\bin'
Add-ToUserPath $goBin

# cgo must be enabled; the build links C code. Set it both as a user env var and
# in Go's own env file, since a stale "go env -w CGO_ENABLED=0" would otherwise
# silently disable cgo and skip the C link entirely.
[Environment]::SetEnvironmentVariable('CGO_ENABLED', '1', 'User')
Write-Ok "Set CGO_ENABLED=1 (user)"
$goExe = Get-Command go -ErrorAction SilentlyContinue
if ($goExe) {
    & $goExe.Source env -w CGO_ENABLED=1
    Write-Ok "Set CGO_ENABLED=1 (go env file)"
}

# Make this very session usable too, so the verification below works.
$env:Path = "$mingwBin;$goBin;$env:Path"
$env:CGO_ENABLED = '1'

# --- Verify ----------------------------------------------------------------

Write-Step "Verifying toolchain"

$gccVersion = (& $gccExe --version | Select-Object -First 1)
Write-Ok "gcc : $gccVersion"

$goCmd = Get-Command go -ErrorAction SilentlyContinue
if ($goCmd) {
    $goVersion = (& go version)
    Write-Ok "go  : $goVersion"
} else {
    Write-Warn2 "go not on PATH in this session yet (it will be in a new terminal)."
}

$gitCmd = Get-Command git -ErrorAction SilentlyContinue
if ($gitCmd) {
    Write-Ok "git : $(& git --version)"
} else {
    Write-Warn2 "git not on PATH in this session yet (it will be in a new terminal)."
}

Write-Host ""
Write-Host "Setup complete." -ForegroundColor Green
Write-Host "Open a NEW Git Bash window and build with:" -ForegroundColor Green
Write-Host "    ./build.sh           # build" -ForegroundColor Green
Write-Host "    ./build.sh --all     # build + lint + race tests" -ForegroundColor Green
Write-Host ""
Write-Host "Do NOT use the MSYS2 shell to build; Git Bash now sees the mingw-w64 gcc on PATH." -ForegroundColor Yellow
