# IDUNA.GAME — CLAUDE.md

## What this is

A real fork of `PITVIPER` (SDL2 terminal emulator, Go), reskinned into Solarized Light with a
text-based prompt sequence that shows IDUNA's real honor code and walks the user through IDUNA's
real device-auth login flow. Launched as a third app from `SHANKPIT/apps/lobby`'s grid, alongside
DEADWEIGHT and PITVIPER. See `NORTHSTAR.md` for the full real investigation this was built on.

**PITVIPER itself is untouched** — same "new fork, not rewrite-in-place" precedent SAND set
against PITVIPER. Only `cmd/idunagame/main.go`'s palette and two new files
(`cmd/idunagame/iduna_prompt.go`, `internal/reflux/`) are IDUNA.GAME's own.

## Build & run

```sh
sudo apt-get install -y libsdl2-dev libsdl2-ttf-dev
make                                              # builds ./idunagame
IDUNA_HOST=127.0.0.1 IDUNA_PORT=8080 ./idunagame  # env vars, default 127.0.0.1:8080
```

Not part of the root `go.work` workspace — always build with `GOWORK=off` (the Makefile already
does this).

## What's real, checked directly (not assumed)

- **Solarized Light palette**: `cmd/idunagame/main.go`'s `baseColors`/`idunaPalette` — real,
  canonical Ethan Schoonover values, not tinted.
- **The honor code**: `cmd/idunagame/honor_code_text.go`, bundled verbatim from
  `IDUNA/internal/honorcode/honorcode.go`'s own `CurrentText`/`CurrentVersion`.
- **Device-auth login**: `cmd/idunagame/iduna_prompt.go` — real HTTP calls to a real IDUNA
  instance's `/auth/device/start`, `/auth/device/poll`, `/auth/token/exchange`. Live-verified
  against this box's own running IDUNA (`:8080`) with a real screenshot showing real honor-code
  text and a real device code rendered in Solarized colors.
- **REFLUX**: `internal/reflux/` — a real, native Go port of SHANKPIT's own
  `packages/reflux/reflux_runtime.c` (same ring-buffer API/ABI, ported not reinvented). Dispatches
  `ActionHonorCodeShown`/`ActionDeviceLinkStarted`/`ActionDeviceLinked` at real milestones. No
  subscriber exists yet — real, by design, matching REFLUX's own "dispatcher never needs to know
  who's listening" model.

## What's NOT done (named, not hidden — see NORTHSTAR.md §4)

- No embedded Google OAuth (not possible for any native app — device auth is the real bridge).
- No `/me`/`/honor-code/accept` calls (real ES256-vs-HS256 token mismatch found and documented,
  not worked around — the exchange response's own inlined `me` object is used instead).
- Shell access isn't gated behind honor-code completion yet — the prompt runs concurrently with a
  normal shell, not before it. Real, deeper event-loop surgery needed for true gating.
- PARENA editor font rendering (`stdlib/editor/render.prn`, same cgo pattern
  `internal/scrollmod/vterm_mod.prn` already proves) — named as a real next step, not started this
  pass; this app currently renders text through PITVIPER's own existing Go font stack
  (`internal/font/`), not PARENA's.
- A visual redraw artifact was observed in the one live screenshot taken so far (partial dark
  bands on some terminal rows) — likely a mid-frame capture timing artifact, not confirmed as a
  real rendering bug; worth a closer look before calling the visuals fully polished.

## Apple Filing Protocol

```bash
emily apples post -t completion -repo IDUNA.GAME "<title>" "<body with commit hash>"
```

## CHANGELOG Protocol

Append a dated bullet to `CHANGELOG.md` for any meaningful change.

## Golden Doc Registration

`NORTHSTAR.md` is registered in `EMILY/context/golden-docs-index.md`.

## Related Repos

- `PITVIPER` — the repo this was forked from; untouched by this fork.
- `IDUNA` — the real backend this app talks to (`/auth/device/*`, honor code source of truth).
- `SHANKPIT` — hosts the lobby app-launcher this app is wired into (`apps/lobby/src/main.c`).
- `PARENA` — `stdlib/editor/render.prn` is the real, not-yet-wired-in font-rendering target.

## Commit Protocol (standing instruction, monorepo-wide)

Always commit and push completed work immediately. Every commit ends with a blank line then
`session: <tag>` (`emily session current`).
