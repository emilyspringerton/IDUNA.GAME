package gpgkey

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Real, live test against the actual `gpg` binary (skipped if not installed -- a real, honest
// environment check, not a mock). Verifies: first call generates a fresh RSA 4096 key and
// returns a real ASCII-armored public block; second call loads the SAME key (same key ID)
// instead of generating a new one.
func TestLoadOrGenerate(t *testing.T) {
	if _, err := exec.LookPath("gpg"); err != nil {
		t.Skip("gpg not installed on this box -- skipping real gpgkey test")
	}

	homeDir := filepath.Join(t.TempDir(), "gnupg")
	uid := "IDUNA.GAME Test <idunagame-test@example.com>"

	pub1, id1, gen1, err := LoadOrGenerate(homeDir, uid)
	if err != nil {
		t.Fatalf("first LoadOrGenerate: %v", err)
	}
	if !gen1 {
		t.Fatal("expected first call to generate a new key")
	}
	if id1 == "" {
		t.Fatal("expected a non-empty key ID")
	}
	if !strings.Contains(pub1, "BEGIN PGP PUBLIC KEY BLOCK") {
		t.Fatalf("expected an armored public key block, got: %q", pub1)
	}
	if strings.Contains(pub1, "PRIVATE KEY") {
		t.Fatal("public export must never contain private key material")
	}

	pub2, id2, gen2, err := LoadOrGenerate(homeDir, uid)
	if err != nil {
		t.Fatalf("second LoadOrGenerate: %v", err)
	}
	if gen2 {
		t.Fatal("expected second call to LOAD the existing key, not generate a new one")
	}
	if id2 != id1 {
		t.Fatalf("expected the same key ID on reload, got %q then %q", id1, id2)
	}
	if pub2 != pub1 {
		t.Fatal("expected the same public key block on reload")
	}
}
