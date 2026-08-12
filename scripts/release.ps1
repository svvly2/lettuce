param([Parameter(Mandatory=$true)][ValidatePattern('^\d+\.\d+\.\d+$')][string]$Version)
$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$desktop = Join-Path $repo 'desktop'

if (-not $env:GH_TOKEN) { throw 'Set GH_TOKEN to a GitHub token with Contents: write permission.' }
& (Join-Path $PSScriptRoot 'check-secrets.ps1')
npm.cmd version $Version --no-git-tag-version --prefix $desktop
rojo build (Join-Path $repo 'plugin\default.project.json') -o (Join-Path $desktop 'Lettuce-Plugin.rbxm')
$env:GOCACHE = Join-Path $repo '.gocache'
$env:GOMODCACHE = Join-Path $repo '.gomodcache'
go build -o (Join-Path $desktop 'lettuce-daemon.exe') (Join-Path $repo 'cmd\daemon')
npm.cmd run release:win --prefix $desktop
Write-Host "Published Lettuce $Version. Commit package.json/package-lock.json and tag v$Version."
