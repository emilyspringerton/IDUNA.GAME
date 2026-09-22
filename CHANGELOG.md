# Changelog

## 2026-09-22 (3)

- New in-app affordance to generate/view this device's own GPG (OpenPGP) signing key (founder
  real-time, direct follow-up to the SSH key affordance above: "now we need gpg key generation
  same thing"). New `internal/gpgkey` — shells out to the real, already-installed `gpg` binary
  (RSA 4096, sign-capable, no passphrase) rather than reimplementing OpenPGP in Go, into a
  self-contained keyring (`~/.config/idunagame/gnupg`, never the caller's own default `~/.gnupg`).
  `iduna_prompt.go`'s new `showDeviceGpgKey` runs right after `showDeviceKey`, printing the
  real ASCII-armored public key block; private material never leaves the device, same as the SSH
  key. Real motivating case: DEADWEIGHT's own app-release signing key (`IDUNA/docs/
  APP_RELEASE_SIGNING.md`) exists only because someone ran a bare `gpg --gen-key` by hand once --
  this makes producing one repeatable and in-app. New `reflux.ActionGpgKeyReady` milestone.
  Desktop-only (no `gpg` on Android) -- named, not silently skipped. Real, live, passing test
  against the actual gpg binary (`internal/gpgkey/gpgkey_test.go`): first run generates a fresh
  key, second run loads the same key ID and identical public block.

## 2026-09-22 (2)

- New in-app affordance to generate/view this device's own Ed25519 SSH key (founder real-time:
  "we need to build the affordances for me to generate the keys into the shankpit client itself...
  we can add that to the iduna app in shankpit"). The keygen mechanism (`internal/sshkey`) already
  existed but was only reachable behind the `-ssh user@host` CLI flag; `iduna_prompt.go`'s own
  `showDeviceKey` now runs it as a real step of the login sequence, right after a device links,
  printing the public key line into the terminal. New `reflux.ActionKeyReady` milestone. Real,
  passing test (`showdevicekey_smoke_test.go`, HOME-isolated): first run generates a fresh key,
  second run loads the same one rather than regenerating.

## 2026-09-22

- New repo, forked from PITVIPER (founder real-time: "have it be a reskin of pitviper into
  solarized light with a text based prompt system"). Real Solarized Light palette swap
  (`cmd/idunagame/main.go`), a text-based IDUNA prompt sequence showing the real honor code and
  walking through IDUNA's real device-auth login flow (`cmd/idunagame/iduna_prompt.go`), and a
  native Go port of SHANKPIT's REFLUX pub/sub event log (`internal/reflux/`). Live-verified
  against a real running IDUNA instance (`:8080`) with a real screenshot. See `NORTHSTAR.md`.
