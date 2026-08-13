param([Parameter(Mandatory=$true)][ValidatePattern('^\d+\.\d+\.\d+$')][string]$Version)
$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$desktop = Join-Path $repo 'desktop'

if (-not $env:GH_TOKEN) { throw 'Set GH_TOKEN to a GitHub token with Contents: write permission.' }
& (Join-Path $PSScriptRoot 'check-secrets.ps1')
npm.cmd version $Version --allow-same-version --no-git-tag-version --prefix $desktop
rojo build (Join-Path $repo 'plugin\default.project.json') -o (Join-Path $desktop 'Lettuce-Plugin.rbxm')
$env:GOCACHE = Join-Path $repo '.gocache'
$env:GOMODCACHE = Join-Path $repo '.gomodcache'
go build -o (Join-Path $desktop 'lettuce-daemon.exe') (Join-Path $repo 'cmd\daemon')
npm.cmd run package:win --prefix $desktop
$tag = "v$Version"
$assets = @(
    (Join-Path $desktop "release\Lettuce-$Version-x64.exe"),
    (Join-Path $desktop "release\Lettuce-$Version-x64.exe.blockmap"),
    (Join-Path $desktop 'release\latest.yml'),
    (Join-Path $desktop 'Lettuce-Plugin.rbxm')
)
gh release create $tag $assets --verify-tag --title "Lettuce $Version" --notes 'See the in-app update log and CHANGELOG.md for everything included in this release.'
Write-Host "Published Lettuce $Version."
