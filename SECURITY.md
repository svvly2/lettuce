# Security policy

## Reporting

Do not open a public issue for token exposure, authentication bypasses, unsafe localhost behavior, or update-channel vulnerabilities. Use GitHub's private vulnerability reporting for this repository.

Include the affected version, reproduction steps, impact, and any relevant logs with credentials removed.

## Credential model

Lettuce is an OAuth 2.0 PKCE public client. The Roblox Client ID is public; Lettuce does not accept or ship a client secret. OAuth tokens are stored in the operating-system credential vault. GitHub publishing tokens and code-signing certificates must only exist in protected repository secrets.

Run `scripts/check-secrets.ps1` before every release. Revoke any credential immediately if it appears in source, logs, chat, or a release artifact.
