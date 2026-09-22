// iduna_prompt.go -- the real, new thing this fork adds on top of PITVIPER's own engine
// (NORTHSTAR.md §3): a text-based prompt sequence, written directly into the vterm the same way
// this file's own sibling main.go already writes its "No SSH key found" notice, that shows the
// real honor code and walks the user through IDUNA's real device-auth login flow.
//
// Runs as a background goroutine (see main.go's own `go runIdunaPrompt(screen)` call, right after
// the real shell/SSH/MUD connection is established) so the normal terminal keeps rendering and
// usable the whole time -- a real, honest v0 tradeoff, not a fully gated "no shell until honor
// code accepted" flow (NORTHSTAR.md §4 names that as real follow-up, not attempted here: doing it
// properly needs intercepting keyboard-to-PTY routing during a pre-shell phase, a deeper change
// to main.go's own event loop than this pass makes).
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"idunagame/internal/gpgkey"
	"idunagame/internal/reflux"
	"idunagame/internal/sshkey"
	"idunagame/internal/vterm"
)

const (
	idunaHostDefault = "127.0.0.1"
	idunaPortDefault = "8080"
)

func idunaHostPort() (string, string) {
	host := os.Getenv("IDUNA_HOST")
	if host == "" {
		host = idunaHostDefault
	}
	port := os.Getenv("IDUNA_PORT")
	if port == "" {
		port = idunaPortDefault
	}
	return host, port
}

type deviceStartResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type devicePollResponse struct {
	Status       string `json:"status"`
	Interval     int    `json:"interval,omitempty"`
	ExchangeCode string `json:"exchange_code,omitempty"`
}

type exchangeMe struct {
	ID     string `json:"id"`
	Handle string `json:"handle"`
	Status string `json:"status"`
}

type exchangeResponse struct {
	AccessToken string     `json:"access_token"`
	Me          exchangeMe `json:"me"`
}

func idunaPost(path string, body any) (*http.Response, error) {
	host, port := idunaHostPort()
	url := fmt.Sprintf("http://%s:%s%s", host, port, path)
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 8 * time.Second}
	return client.Post(url, "application/json", bytes.NewReader(buf))
}

// solGold writes text into the terminal using ANSI 256-color 3 (Solarized "yellow" in
// idunagame's own reskinned baseColors table, main.go) -- IDUNA's own accent, same real reasoning
// idunaPalette.accent's own doc comment gives for that color choice.
func solGold(s string) string { return "\x1b[33m" + s + "\x1b[0m" }
func solMuted(s string) string { return "\x1b[36m" + s + "\x1b[0m" }
func solBold(s string) string  { return "\x1b[1m" + s + "\x1b[0m" }

// runIdunaPrompt is the real sequence: show the honor code, start the real device-auth flow
// (NORTHSTAR.md §3), poll it to completion, and dispatch real REFLUX events (founder real-time:
// "build in REFLUX") at each real milestone so any other IDUNA.GAME subscriber -- there are none
// yet, same "dispatcher never needs to know who's listening" REFLUX design reflux_runtime.h's own
// doc comment describes -- can pick them up later with zero coupling to this file.
func runIdunaPrompt(screen *vterm.Screen) {
	screen.Write([]byte("\r\n" + solBold("THE HONOR CODE") + fmt.Sprintf(" (v%d)\r\n\r\n", honorCodeVersion)))
	for _, line := range splitLines(honorCodeText) {
		screen.Write([]byte(line + "\r\n"))
	}
	reflux.Dispatch(reflux.ActionHonorCodeShown, honorCodeVersion, 0, 0)

	screen.Write([]byte("\r\n" + solMuted("Requesting a device link code from IDUNA...") + "\r\n"))

	startResp, err := idunaPost("/auth/device/start", map[string]any{})
	if err != nil {
		screen.Write([]byte(solMuted(fmt.Sprintf("Could not reach IDUNA: %v\r\n", err))))
		return
	}
	defer startResp.Body.Close()
	var start deviceStartResponse
	if err := json.NewDecoder(startResp.Body).Decode(&start); err != nil || start.DeviceCode == "" {
		screen.Write([]byte(solMuted("IDUNA did not return a real device code -- device auth may not be configured here.\r\n")))
		return
	}
	reflux.Dispatch(reflux.ActionDeviceLinkStarted, 0, 0, 0)

	screen.Write([]byte("\r\n" + solBold("LINK THIS DEVICE") + "\r\n"))
	screen.Write([]byte("  1. On any browser, visit: " + solGold(start.VerificationURL) + "\r\n"))
	screen.Write([]byte("  2. Sign in, accept the honor code, and enter this code: " + solGold(start.UserCode) + "\r\n\r\n"))

	interval := start.Interval
	if interval <= 0 {
		interval = 2
	}

	deadline := time.Now().Add(time.Duration(start.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(time.Duration(interval) * time.Second)

		pollResp, err := idunaPost("/auth/device/poll", map[string]any{"device_code": start.DeviceCode})
		if err != nil {
			continue // real, transient network hiccup -- keep polling, same cadence
		}
		var poll devicePollResponse
		decodeErr := json.NewDecoder(pollResp.Body).Decode(&poll)
		pollResp.Body.Close()
		if decodeErr != nil {
			continue
		}
		if pollResp.StatusCode == http.StatusTooManyRequests {
			continue // Retry-After respected implicitly by our own interval cadence
		}
		if poll.Status != "authorized" {
			continue
		}

		exResp, err := idunaPost("/auth/token/exchange", map[string]any{"exchange_code": poll.ExchangeCode})
		if err != nil {
			screen.Write([]byte(solMuted(fmt.Sprintf("Token exchange failed: %v\r\n", err))))
			return
		}
		var exchange exchangeResponse
		exDecodeErr := json.NewDecoder(exResp.Body).Decode(&exchange)
		exResp.Body.Close()
		if exDecodeErr != nil || exResp.StatusCode != http.StatusOK {
			screen.Write([]byte(solMuted("Device was authorized but the token exchange failed -- " +
				"if you just accepted the honor code, try linking again.\r\n")))
			return
		}

		handle := exchange.Me.Handle
		if handle == "" {
			handle = "(no handle)"
		}
		screen.Write([]byte("\r\n" + solBold("LINKED") + " -- welcome, " + solGold(handle) + ".\r\n\r\n"))
		reflux.Dispatch(reflux.ActionDeviceLinked, 1, 0, 0)
		showDeviceKey(screen)
		showDeviceGpgKey(screen, handle)
		return
	}
	screen.Write([]byte(solMuted("Device link code expired before it was authorized -- restart idunagame to try again.\r\n")))
}

// showDeviceKey is the real, in-app affordance for generating (or loading) this device's own
// Ed25519 SSH keypair -- founder real-time: "we need to build the affordances for me to generate
// the keys into the shankpit client itself... we can add that to the iduna app in shankpit."
// The keygen mechanism itself already existed (internal/sshkey, ported from PITVIPER's own
// 2026-09-06 "the phone app will need a way to generate the key" work) but was only ever reachable
// behind the `-ssh user@host` CLI flag in main.go -- never a real, discoverable affordance inside
// the app itself. This surfaces it as a first-class step of IDUNA.GAME's own login sequence,
// right after a device links, using the exact same keyPath convention main.go's own -ssh flow
// already establishes (~/.config/idunagame/id_ed25519 on Linux, the platform-equivalent
// UserConfigDir elsewhere) so both code paths always load/generate the same one real key --
// never a second, divergent keypair. Same real, deliberate design sshkey's own doc comment
// names: the PRIVATE half is generated on-device and never leaves it; only the PUBLIC line is
// ever shown here, safe to copy/share/photograph/QR.
func showDeviceKey(screen *vterm.Screen) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		cfgDir = "."
	}
	keyPath := filepath.Join(cfgDir, "idunagame", "id_ed25519")
	pubLine, generated, err := sshkey.LoadOrGenerate(keyPath)
	if err != nil {
		screen.Write([]byte(solMuted(fmt.Sprintf("Could not generate a device key: %v\r\n", err))))
		return
	}
	genFlag := int32(0)
	if generated {
		genFlag = 1
	}
	reflux.Dispatch(reflux.ActionKeyReady, genFlag, 0, 0)

	if generated {
		screen.Write([]byte(solBold("YOUR DEVICE KEY") + " -- generated a new one, saved to " + keyPath + "\r\n"))
	} else {
		screen.Write([]byte(solBold("YOUR DEVICE KEY") + " -- loaded from " + keyPath + "\r\n"))
	}
	screen.Write([]byte("The private half never leaves this device. Use the public line below\r\n"))
	screen.Write([]byte("anywhere you need to authenticate as this device (e.g. as a remote\r\n"))
	screen.Write([]byte("host's ~/.ssh/authorized_keys entry, via -ssh user@host):\r\n\r\n"))
	screen.Write([]byte("  " + solGold(pubLine) + "\r\n\r\n"))
}

// showDeviceGpgKey is GPG's own real counterpart to showDeviceKey above -- founder real-time,
// direct follow-up: "now we need gpg key generation same thing". Real, live motivating case:
// DEADWEIGHT's own app-release signing pipeline (IDUNA/docs/APP_RELEASE_SIGNING.md) needs a real
// RSA 4096 GPG signing key, which today only ever existed because someone ran a bare `gpg
// --gen-key` by hand once -- this makes producing one a real, repeatable, in-app affordance
// instead. Same real design as showDeviceKey: a self-contained keyring (never the caller's own
// default ~/.gnupg), only the PUBLIC armored block is ever shown, and a second run loads the
// same key rather than generating a new one every time. handle is the real, just-linked IDUNA
// handle from the device-auth exchange -- used as the GPG identity's own uid (no real email is
// available from that exchange response, so a synthetic-but-real, stable per-handle address is
// used instead, same spirit as any local-identity uid that doesn't need to receive mail).
func showDeviceGpgKey(screen *vterm.Screen, handle string) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		cfgDir = "."
	}
	homeDir := filepath.Join(cfgDir, "idunagame", "gnupg")
	uid := fmt.Sprintf("%s <%s@idunagame.local>", handle, handle)

	pubArmored, keyID, generated, err := gpgkey.LoadOrGenerate(homeDir, uid)
	if err != nil {
		if err == gpgkey.ErrGPGNotInstalled {
			screen.Write([]byte(solMuted("No gpg binary found on this device -- GPG key generation needs " +
				"a real GnuPG install (desktop only; not available on Android yet).\r\n\r\n")))
			return
		}
		screen.Write([]byte(solMuted(fmt.Sprintf("Could not generate a GPG key: %v\r\n", err))))
		return
	}
	genFlag := int32(0)
	if generated {
		genFlag = 1
	}
	reflux.Dispatch(reflux.ActionGpgKeyReady, genFlag, 0, 0)

	if generated {
		screen.Write([]byte(solBold("YOUR GPG KEY") + " -- generated a new one (RSA 4096), keyring at " + homeDir + "\r\n"))
	} else {
		screen.Write([]byte(solBold("YOUR GPG KEY") + " -- loaded from " + homeDir + "\r\n"))
	}
	screen.Write([]byte("Key ID: " + solGold(keyID) + "\r\n"))
	screen.Write([]byte("The private half never leaves this device. Public key (paste anywhere\r\n"))
	screen.Write([]byte("you need to publish/verify signatures, e.g. a GitHub Actions secret):\r\n\r\n"))
	for _, line := range splitLines(pubArmored) {
		screen.Write([]byte("  " + line + "\r\n"))
	}
	screen.Write([]byte("\r\n"))
}

// splitLines is a tiny, dependency-free \n splitter -- honorCodeText's own real newlines, same
// shape strings.Split(s, "\n") gives, kept local to avoid importing "strings" for one call site.
func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}
