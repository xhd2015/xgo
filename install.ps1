#!/usr/bin/env pwsh

if (!(Get-Command go -errorAction SilentlyContinue))
{
    Write-Output "requires go to install xgo"
    exit 1
}

$ErrorActionPreference = "Stop" 

$XgoRoot = "${Home}\.xgo" 
$XgoBin = mkdir -Force "${XgoRoot}\bin"
$XgoInstallSrc = mkdir -Force "${XgoRoot}\install-src"

Remove-Item -Recurse -Force "${XgoInstallSrc}"

$URL = "https://github.com/xhd2015/xgo/releases/download/v1.0.6/install-src.zip"
$ZipPath = "${XgoBin}\install-src.zip"

# curl.exe is faster than PowerShell 5's 'Invoke-WebRequest'
# note: 'curl' is an alias to 'Invoke-WebRequest'. so the exe suffix is required
curl.exe "-#SfLo" "$ZipPath" "$URL"

Expand-Archive "$ZipPath" "${XgoInstallSrc}" -Force

cd "${XgoInstallSrc}"

go run "./"