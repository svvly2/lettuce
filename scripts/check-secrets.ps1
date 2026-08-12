$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent $PSScriptRoot
$excluded = @('desktop\node_modules\', 'desktop\release\', '.gomodcache\', '.gocache\')
$patterns = @(
    'RBX-[A-Za-z0-9_-]{16,}',
    '-----BEGIN (RSA |EC |OPENSSH )?PRIVATE KEY-----',
    '(GH_TOKEN|GITHUB_TOKEN)\s*=\s*[^$\s][^\s]*',
    'oauth_client_secret\s*=\s*\S+'
)
$hits = Get-ChildItem -LiteralPath $repo -Recurse -File | Where-Object {
    $relative = $_.FullName.Substring($repo.Length + 1)
    $_.FullName -ne $PSCommandPath -and -not ($excluded | Where-Object { $relative.StartsWith($_) })
} | Select-String -Pattern $patterns
if ($hits) {
    $hits | ForEach-Object { Write-Error "Possible secret: $($_.Path):$($_.LineNumber)" }
    exit 1
}
Write-Host 'Secret scan passed.'
