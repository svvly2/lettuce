# Changelog

## 0.1.17

- Automatically removed completed and failed queue items after 30 seconds.
- Preserved newer retries when cleanup timers from older jobs finish.
- Added smooth visual transitions for changing queue states.
- Replaced concurrent Electron publishing with one atomic GitHub release upload.

## 0.1.16

- Made failed uploads immediately visible with a consistent red error state.
- Kept Queue and Activity at a fixed size regardless of item count.
- Added dedicated, natural scrolling inside Queue and Activity lists.
- Prevented animation and activity content from scrolling or resizing the main page.
- Added a clean, expandable in-app history of previous updates.

## 0.1.15

- Added a compact, once-per-version release note on the main page.
- Kept Settings exclusively in the bottom navigation.
- Simplified the account menu and refined the Studio port typography.
- Simplified the public project page and download flow.
- Moved automatic update scheduling into a focused desktop module.
- Added clean GitHub CI, release automation, and secret scanning.

## 0.1.14

- Added Roblox OAuth 2.0 Authorization Code with PKCE.
- Added secure token persistence and automatic refresh.
- Added OAuth asset-delivery and Create Asset uploads for animations and sounds.
- Rebuilt the Studio plugin with automatic discovery, replacement, responsive UI, and bottom navigation.
- Added the Lettuce logo and refined desktop UI/icons.
- Added NSIS automatic updates through GitHub Releases.
- Added release-time secret scanning and removed client-secret support entirely.

## 0.1.0

- Initial Lettuce desktop and Studio integration.
