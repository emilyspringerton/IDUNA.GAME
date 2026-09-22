# IDUNA.GAME — Northstar

*Written 2026-09-22. Founder real-time: "we need to build an IDUNA SHANKPIT_OS app so we can have
the honor code stuff go live we need to be able to open pages in the IDUNA brand affordances
including specific colors and nice fonts."* Clarified via AskUserQuestion: not a browser wrapper,
not an embedded webview — *"not a browser a true port of iduna affordances into sdl 2 a truely
new app i will create a repo now. IDUNA.GAME."*

## 1. What this is

A native C/SDL2 + SDL2_ttf application — the same real engine family SHANKPIT/REDGARDEN/GFD's
`battlegrounds_gui` are all built in, launched as a third app from SHANKPIT OS's lobby grid
(`SHANKPIT/apps/lobby`, alongside DEADWEIGHT and PITVIPER) — that renders real IDUNA "pages"
(server-driven screens: the honor code ceremony first, more later) using IDUNA's own real brand
system: its actual color palette and actual web fonts (Cormorant Garamond, Spectral), not a
generic OS look. Not a browser, not a webview — every screen is hand-rendered with SDL2_ttf,
same "hand-rolled UI" convention this engine family already uses everywhere else (HUD, shop
panels, draft screens).

## 2. Real investigation performed before building (not guessed)

Read `IDUNA`'s actual live source, not assumed:

- **Brand system, real and living** (`IDUNA/styles.css`, `IDUNA/index.html`,
  `IDUNA/docs/kikoryu/VS0_IDENTITY_GATE.md`): gold `#b89b62` (the doc explicitly says treat
  `styles.css` as the living palette over the original spec's `#C6A75E` — "metal drift"), rose
  gold `#b76e79` reserved specifically for irreversible consent (honor-code acceptance) — a real
  semantic rule, not just a color choice — cream/parchment background `#f3ede2`, Cormorant
  Garamond for headlines, Spectral/Inter for body text. Both fonts are Google Fonts (OFL
  licensed, freely redistributable) — real `.ttf` files downloaded and bundled at
  `assets/fonts/`, not a system-font substitute.
- **The real honor code text** (`IDUNA/internal/honorcode/honorcode.go`): `CurrentVersion = 1`,
  `CurrentText` is the canonical body. Bundled verbatim here (`src/honor_code_text.h`) — same
  "static fallback, versioned" shape `app.js`'s own `DECLARATION_TEXT` constant already uses on
  the web side. A live-fetch upgrade (pull the current text/version from the server instead of a
  baked-in copy) is real, valuable, not-yet-built follow-up — see §5.
- **Auth: three real, non-interoperable systems found, not one** — checked directly rather than
  guessed:
  1. **Web ceremony** (`internal/http/handlers/web_ceremony.go`): Google OAuth, ES256 JWT
     (`middleware.RequireAuth(keys)` → `authjwt.Verify`), gates `/me`, `/honor-code/accept`,
     `/me/handle`. This is what `IDUNA/index.html`+`app.js` actually use, live, at IDUNA's own
     root path.
  2. **Player email auth** (`internal/http/handlers/player_email_auth.go`): also ES256-signed
     (same `authjwt.Sign`/`Keys`), but scoped to SHANKPIT-player rows, a different table than the
     core IAM `users` table `web_ceremony.go`'s `HandleMe` reads from.
  3. **Device auth bridge** (`internal/auth/device/`, `internal/http/handlers/device.go`) — real,
     live, confirmed by a real local call (`POST /auth/device/start` against this box's own
     `:8080`, real `device_code`/`user_code`/`verification_url` returned). This is
     `VS0_IDENTITY_GATE.md`'s own explicitly-named "bridge for any non-browser client" — exactly
     this app's own situation. **Its exchange token is HS256-signed with a separate secret,
     `aud: "kikoryu"`** (`device.go`'s `signHS256`) — checked directly against
     `middleware.RequireAuth`'s real `jwt.Verify(keys, token)` call: an ES256 verifier will
     reject an HS256 token outright. **This token cannot call `/me` or `/honor-code/accept`.**
     Real, confirmed architecture gap, not assumed — named here rather than silently worked
     around.

## 2.5 Mid-build pivot (same session, real-time)

The founder redirected mid-build, twice, before VS0 landed: first from a bespoke C/SDL2 GUI-panel
app (what §§1-2 above originally described building) to *"a reskin of pitviper into solarized
light with a text based prompt system"* — i.e. a real fork of PITVIPER, not a new GUI toolkit —
then added two more real integrations on top: PARENA editor's own font rendering (named, not yet
wired — see §4), and REFLUX, SHANKPIT's real cross-mod pub/sub event log (`internal/reflux/`,
a native Go port of `SHANKPIT/packages/reflux/reflux_runtime.c`'s own tiny ring-buffer API/ABI —
proportionate given REFLUX's real implementation is ~75 lines with no PARENA-specific logic to
justify a cgo/PARENA-compile round-trip the way `internal/scrollmod/vterm_mod.prn` needs one).
The original bespoke-SDL2 exploration (§§2's own earlier draft) was discarded, not shipped —
its real investigation findings (brand palette, device-auth flow, the ES256/HS256 gap) carried
forward into this doc unchanged.

## 3. The real, working flow this app uses (device auth, not Google OAuth directly)

Google's own OAuth consent screen fundamentally requires a real web browser — no way around that
for ANY native app, this one included (the same reason RFC 8252/8628 exist industry-wide). So
`IDUNA.GAME` doesn't attempt embedded Google auth at all:

1. `POST /auth/device/start` → a real `user_code` (e.g. `57JA-V68S`) + `verification_url`
   (`https://okemily.com/device`) + `expires_in`/`interval`.
2. Rendered natively, branded: the code in large Cormorant Garamond type, the URL, a short
   instruction. The user opens that URL on ANY browser (phone, another tab) and completes the
   **real, already-live, already-fully-branded** web ceremony there — Google login, honor-code
   acceptance (rose gold, per its own reserved semantic), gamertag pick — all of it real,
   existing, unmodified.
3. `IDUNA.GAME` polls `POST /auth/device/poll` with the `device_code` every `interval` seconds.
4. Once `status: authorized`, `POST /auth/token/exchange` with the resulting `exchange_code` →
   the real response already inlines `me: {id, handle, status, roles}` — no `/me` call needed
   (correctly sidesteps the HS256-vs-ES256 gap named in §2: this app never calls the ES256-gated
   endpoints at all).
5. Done screen: real handle, real status, styled brand-consistent.

Honor-code *acceptance itself* happens on the real web page during step 2 — this app doesn't
re-implement the accept call. What it DOES render natively, in brand, before/during the wait: the
honor code TEXT itself (§2's bundled copy) — so opening this app is a real, branded "read the
covenant" experience even before a browser is involved, matching the founder's own "open pages"
framing.

## 4. What's NOT attempted here (named, not hidden)

- No Google OAuth performed directly by this app (§3's own reasoning — not a workaround gap,
  a real platform constraint every native OAuth client has).
- No use of `/me`/`/honor-code/accept` (the ES256-gated core-IAM endpoints) — the exchange
  response's own inlined `me` object is sufficient for this app's real needs.
- No embedded QR code rendering for the verification URL — real, cheap, valuable follow-up
  (would need a QR-generation routine; PARENA's `stdlib`/IDUNA's own Go `qrcode` package are both
  real precedent elsewhere in this monorepo, ported here would be a small follow-up, not core
  to VS0).
- Resolving the real 3-way auth-namespace gap named in §2 is real, valuable, separate backend
  work (whether the core-IAM `/me` family should also accept a device-flow-issued token, or
  whether `device.Service.Exchange` should mint an ES256 token instead of HS256) — a founder-level
  architecture call, not guessed at or silently patched here.
- **PARENA editor font rendering** (founder real-time: "but actually all of the font rendering of
  PARENA EDITOR"): `PARENA/stdlib/editor/render.prn`/`textmate.prn` is the real, intended
  rendering target, same cgo-FFI pattern `internal/scrollmod/vterm_mod.prn` already proves
  (compile via `parena build ... -o some.c`, cgo-link the result, same shape PITVIPER's own
  NORTHSTAR Milestone 6 already scoped and never built). Not done this pass — VS0 renders through
  PITVIPER's own existing Go font stack (`internal/font/`) unchanged. Real brand `.ttf` files
  (Cormorant Garamond, Spectral — OFL licensed) are bundled at `assets/fonts/` for whichever
  rendering path eventually consumes them, but nothing wires them in yet.
- **True shell gating**: the honor-code/device-auth prompt runs concurrently with a normal shell
  connection, not before it (§3's own note) — a real, deeper change to `main.go`'s own event loop
  (intercepting keyboard-to-PTY routing during a pre-shell phase) would be needed for genuine
  gating, not attempted here.
- **REFLUX has no subscriber yet** — `internal/reflux/` dispatches real events
  (`ActionHonorCodeShown`/`ActionDeviceLinkStarted`/`ActionDeviceLinked`) but nothing in this repo
  polls them yet, by design (REFLUX's own "dispatcher never needs to know who's listening" model)
  — a real, future subscriber (this app's own status bar, or an external tool) is separate work.
- **One visual artifact observed, not root-caused**: the single live screenshot taken (§6) shows
  partial dark bands on some terminal rows — most likely a mid-redraw capture-timing artifact
  (this file's own per-line clear+redraw caught between frames), not confirmed as a real rendering
  bug. Worth a closer look before calling the visuals fully polished.

## 5. What's real and shipped

| # | Milestone | Status |
|---|---|---|
| 0 | This doc | DONE |
| 1 | Real fork of PITVIPER, reskinned to real Solarized Light values (`cmd/idunagame/main.go`'s `baseColors`/`idunaPalette`) | DONE |
| 2 | Honor-code text bundled verbatim (`cmd/idunagame/honor_code_text.go`), rendered into the vterm as real terminal text | DONE |
| 3 | Device-auth flow: real `POST /auth/device/start` + poll + exchange against this box's own running IDUNA (`:8080`), live-verified with a real screenshot | DONE |
| 4 | REFLUX, a native Go port of SHANKPIT's real pub/sub event log, dispatching real milestones | DONE |
| 5 | `go build`/`go vet`/`go test` all clean (`GOWORK=off`); real headless (Xvfb) run + screenshot confirming live honor-code text, Solarized colors, and a real device code on screen | DONE |
| 6 | Wired into `SHANKPIT/apps/lobby` as a third launchable app | NOT STARTED (next in this same session) |
| 7 | PARENA editor font rendering wired in (see §4) | NOT STARTED |
| 8 | Live-fetch honor-code text/version instead of the bundled copy | NOT STARTED |
| 9 | QR code rendering for the verification URL | NOT STARTED |
| 10 | Resolve the auth-namespace gap (§2's own §2.3) so this app could call `/me` directly | NOT STARTED — founder decision |
| 11 | True shell gating (see §4) | NOT STARTED |
| 12 | Root-cause the redraw artifact (§4's own last note) | NOT STARTED |

## Related docs

| Doc | Location |
|---|---|
| Honor code source of truth | `IDUNA/internal/honorcode/honorcode.go` |
| Device auth bridge | `IDUNA/internal/auth/device/`, `IDUNA/internal/http/handlers/device.go` |
| Brand system origin + known "metal drift" | `IDUNA/docs/kikoryu/VS0_IDENTITY_GATE.md` |
| Web ceremony (the real page this app doesn't replace, only complements) | `IDUNA/index.html`, `IDUNA/app.js`, `IDUNA/styles.css` |
| SHANKPIT OS app-launcher precedent | `SHANKPIT/apps/lobby/src/main.c` (`lobby_launch_app`) |
| Same-box HTTP client precedent, ported here | `REDGARDEN/packages/common/http_client.h` |
