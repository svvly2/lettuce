# Lettuce

Lettuce is a lightweight Roblox Studio asset replacement desktop app. The Studio plugin discovers animation references, sends a batch to the loopback service, and applies the returned old-to-new ID mapping. Users never copy animation IDs.

## Architecture

- `desktop/src`: runtime-neutral React UI. It depends only on the `DesktopBridge` interface.
- `desktop/electron`: thin Electron adapter, custom window, system-browser OAuth launch, and daemon lifecycle.
- `cmd/daemon`: local orchestration API and application state.
- `internal/oauth`: OAuth 2.0 Authorization Code + PKCE, refresh, and OS credential-vault persistence.
- `internal/app/assets`: extensible asset handlers and upload pipeline.
- `internal/roblox`: Roblox/Open Cloud clients.
- `plugin`: lightweight Luau Studio client, discovery, and replacement engine.

The distributable Studio plugin is `desktop/release/Lettuce-Plugin.rbxm`. It is
self-contained: its bundled modules only scan and edit the open place and call
Lettuce on `127.0.0.1`. It does not require Asset-Reuploader, a second plugin,
an API module, cookies, API keys, or OAuth tokens in Studio.

The renderer has no Node access. Electron uses context isolation and sandboxing; only a typed, narrow preload bridge is exposed.

## Roblox OAuth configuration

Create a Roblox OAuth 2.0 application and configure:

| Setting | Value |
| --- | --- |
| Flow | Authorization Code with PKCE (`S256`) |
| Client type | Public/native client (no client secret is shipped) |
| Redirect URI | `http://localhost:38073/oauth/callback` |
| Scopes | `openid profile asset:read asset:write` (reduce if your Roblox app supports a smaller set) |
| Authorization endpoint | `https://apis.roblox.com/oauth/v1/authorize` |
| Token endpoint | `https://apis.roblox.com/oauth/v1/token` |
| User info endpoint | `https://apis.roblox.com/oauth/v1/userinfo` |

Set `oauth_client_id`, `oauth_redirect_uri`, and `oauth_scopes` in `config.ini`, or use `ROBLOX_OAUTH_CLIENT_ID`, `ROBLOX_OAUTH_REDIRECT_URI`, and `ROBLOX_OAUTH_SCOPES`. The Client ID is a public identifier. Lettuce is a PKCE public client and has no client-secret setting or code path. Access and refresh tokens are stored under `com.lettuce.desktop / roblox-oauth-session` in the operating-system credential vault. Logout deletes that entry. Tokens refresh shortly before expiry.

The redirect listener binds only to `127.0.0.1`, validates a cryptographically random `state`, requires the original PKCE verifier, expires pending logins, and closes after completion.

## Studio protocol

The plugin talks only to `http://127.0.0.1:38073`:

1. It scans the place for supported animation references.
2. `POST /reupload` sends `{ pluginVersion, placeId, creatorId, isGroup, assetType, ids }`.
3. Lettuce resolves metadata/content, uploads with the current OAuth bearer token, polls long-running operations, retries transient failures with backoff, and emits mapping updates.
4. The plugin applies each returned `{ oldId, newId }` mapping to instances and script sources.
5. The response stream ends with `done`; a disconnect is safe to retry.

Keep Roblox Studio's **Allow HTTP Requests** enabled. The daemon rejects non-loopback traffic and should never be exposed through a router or tunnel.

## Development

```powershell
cd desktop
npm install
npm run start
```

Build the UI with `npm run build`. Build the daemon from the repository root with `go build -o desktop/lettuce-daemon.exe ./cmd/daemon`.

## Shipping automatic updates

Installed Windows builds check the public `svvly2/lettuce` GitHub Releases feed
after launch and every four hours. Updates download in the background and install
when Lettuce exits. The bundled Studio plugin is copied into Roblox's local Plugins
folder on each Lettuce launch, so desktop and plugin updates stay together.

For each release, set a GitHub token with **Contents: write** permission and run:

```powershell
$env:GH_TOKEN = $(Read-Host 'temporary GitHub token')
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\release.ps1 -Version 0.1.15
```

The release must contain the generated installer, blockmap, and `latest.yml`.
Increment the version every time; clients only install a version newer than their
current one. Never commit `GH_TOKEN`. If the repository name changes, update the
`build.publish` owner/repo fields in `desktop/package.json` before publishing.

## Extending asset support

Add a handler behind the existing asset-module boundary, keep Roblox transport code in `internal/roblox`, and add a small plugin-side discovery/replacement adapter. Queueing, retries, logging, OAuth, and UI state remain shared.

## License and reference

The original repository is GPL-3.0. This workspace retains its GPL-3.0 license and attribution. Reference-derived Go and Luau code remains isolated under `internal/`, `cmd/`, and `plugin/`; the new React/Electron desktop layer is a refactored adapter over those boundaries.
