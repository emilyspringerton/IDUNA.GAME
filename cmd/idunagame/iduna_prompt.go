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
	"time"

	"idunagame/internal/reflux"
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
		return
	}
	screen.Write([]byte(solMuted("Device link code expired before it was authorized -- restart idunagame to try again.\r\n")))
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
