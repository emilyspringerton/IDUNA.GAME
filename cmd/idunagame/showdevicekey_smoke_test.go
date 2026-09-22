package main

import (
	"os"
	"strings"
	"testing"

	"idunagame/internal/vterm"
)

// Throwaway smoke test for showDeviceKey -- real HOME-isolated (never touches the developer's
// own ~/.config/idunagame key), verifies it generates a fresh key, writes real terminal output,
// and a second call loads (not regenerates) the same key.
func TestShowDeviceKeySmoke(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/.config")

	screen := vterm.New(80, 24)
	showDeviceKey(screen)
	cells, cols, rows, _, _ := screen.Snapshot()
	var sb strings.Builder
	for _, c := range cells {
		if c.Ch == 0 {
			sb.WriteByte(' ')
		} else {
			sb.WriteRune(c.Ch)
		}
	}
	_ = cols
	_ = rows
	out := sb.String()
	if !strings.Contains(out, "YOUR DEVICE KEY") {
		t.Fatalf("expected YOUR DEVICE KEY banner, got: %q", out)
	}
	if !strings.Contains(out, "generated a new one") {
		t.Fatalf("expected first run to generate a new key, got: %q", out)
	}
	if !strings.Contains(out, "ssh-ed25519") {
		t.Fatalf("expected a real ssh-ed25519 public key line, got: %q", out)
	}

	keyPath := tmp + "/.config/idunagame/id_ed25519"
	if _, err := os.Stat(keyPath); err != nil {
		t.Fatalf("expected key file at %s: %v", keyPath, err)
	}

	screen2 := vterm.New(80, 24)
	showDeviceKey(screen2)
	cells2, _, _, _, _ := screen2.Snapshot()
	var sb2 strings.Builder
	for _, c := range cells2 {
		if c.Ch == 0 {
			sb2.WriteByte(' ')
		} else {
			sb2.WriteRune(c.Ch)
		}
	}
	out2 := sb2.String()
	if !strings.Contains(out2, "loaded from") {
		t.Fatalf("expected second run to LOAD the existing key, not regenerate, got: %q", out2)
	}
}
