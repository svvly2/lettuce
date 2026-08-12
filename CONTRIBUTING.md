# Contributing

Issues and focused pull requests are welcome. Use the repository issue tracker for reproducible bugs and feature proposals, or join the [Lettuce Discord](https://discord.gg/4ycV7TUX6G).

## Local development

Requirements: the Go version declared in `go.mod`, Node.js 22+, Aftman, and Roblox Studio with HTTP requests enabled.

```powershell
go test ./cmd/daemon ./internal/oauth ./internal/roblox/... ./internal/app/assets/animation ./internal/app/assets/sound
cd desktop
npm ci
npm run start
```

Build the plugin from the repository root:

```powershell
rojo build plugin/default.project.json -o desktop/Lettuce-Plugin.rbxm
```

Never commit OAuth tokens, GitHub tokens, API keys, cookies, client secrets, private keys, or signing certificates. Run `scripts/check-secrets.ps1` before opening a pull request.
