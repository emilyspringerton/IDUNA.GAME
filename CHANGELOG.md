# Changelog

## 2026-09-22

- New repo, forked from PITVIPER (founder real-time: "have it be a reskin of pitviper into
  solarized light with a text based prompt system"). Real Solarized Light palette swap
  (`cmd/idunagame/main.go`), a text-based IDUNA prompt sequence showing the real honor code and
  walking through IDUNA's real device-auth login flow (`cmd/idunagame/iduna_prompt.go`), and a
  native Go port of SHANKPIT's REFLUX pub/sub event log (`internal/reflux/`). Live-verified
  against a real running IDUNA instance (`:8080`) with a real screenshot. See `NORTHSTAR.md`.
