// Package gpgkey generates and loads a real, self-contained GPG (OpenPGP) keypair for this
// device, the same real "on-device keygen affordance" role internal/sshkey already plays for SSH.
//
// Founder real-time: "now we need gpg key generation same thing" -- direct follow-up to the SSH
// device key affordance (cmd/idunagame/iduna_prompt.go's own showDeviceKey), itself in response
// to DEADWEIGHT's own real, currently-blocked app-release signing pipeline (IDUNA/docs/
// APP_RELEASE_SIGNING.md): a real RSA 4096 signing key already exists there, but it was generated
// once, by hand, via a bare `gpg --gen-key` -- there was no repeatable, in-app way to produce one.
//
// Real, deliberate design choice: this package SHELLS OUT to the real, already-installed `gpg`
// binary (GnuPG 2.4+, confirmed present) rather than reimplementing OpenPGP keygen in pure Go.
// Same "trust the real, external tool over a from-scratch crypto reimplementation" convention
// this monorepo already follows elsewhere (NOCK's own imagemagick.go shells out to `convert`;
// MIXFORGE's own plan shells out to `yt-dlp`) -- doubly true for cryptography, where a bug in a
// hand-rolled OpenPGP implementation is a real security hole, not just a correctness bug. The one
// real, honest limitation this creates (not silently hidden): `gpg` is a real Linux/Mac/Windows
// desktop binary, not something present on Android -- the same platform gap sshkey.go's own doc
// comment names for its future Android callers, except sshkey.go solved it with pure-Go
// (golang.org/x/crypto/ssh has no external-binary dependency at all); GPG keygen here is
// desktop-only until a real pure-Go OpenPGP library (e.g. github.com/ProtonMail/go-crypto) is
// deliberately adopted for the mobile path -- named here as real, deferred future work, not
// silently assumed away.
//
// Uses a real, self-contained GPG homedir (never the caller's own default ~/.gnupg) so this
// device's own generated identity never mixes with -- or clobbers -- any GPG keyring the person
// running this already has, same isolation sshkey.go's own dedicated id_ed25519 path already
// gives the SSH key.
package gpgkey

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ErrGPGNotInstalled is returned when the `gpg` binary isn't on PATH -- a real, honest, checkable
// precondition (not assumed) the caller should surface to the user rather than a raw exec error.
var ErrGPGNotInstalled = fmt.Errorf("gpgkey: the gpg binary is not installed on this device")

// LoadOrGenerate ensures a real GPG keypair identified by uid (e.g. "Handle <handle@example.com>")
// exists in the self-contained keyring at homeDir, generating one (RSA 4096, sign-capable, no
// expiry, no passphrase -- matching this monorepo's own real app-release signing key's own real
// shape, IDUNA/docs/APP_RELEASE_SIGNING.md) if none exists yet. Returns the ASCII-armored PUBLIC
// key block (safe to display/share/paste into a GitHub Actions secret UI) and the real key ID,
// never the private key material -- same "only the public half is ever surfaced" design
// sshkey.LoadOrGenerate already establishes for SSH.
func LoadOrGenerate(homeDir, uid string) (pubArmored string, keyID string, generated bool, err error) {
	if _, lookErr := exec.LookPath("gpg"); lookErr != nil {
		return "", "", false, ErrGPGNotInstalled
	}
	if mkErr := os.MkdirAll(homeDir, 0o700); mkErr != nil {
		return "", "", false, fmt.Errorf("gpgkey: mkdir %s: %w", homeDir, mkErr)
	}

	id, findErr := findKeyID(homeDir, uid)
	if findErr != nil {
		return "", "", false, findErr
	}
	if id != "" {
		armored, expErr := exportPublic(homeDir, id)
		if expErr != nil {
			return "", "", false, expErr
		}
		return armored, id, false, nil
	}

	// No existing key for this uid -- generate a fresh one. --batch + empty --passphrase +
	// --pinentry-mode loopback is the real, standard non-interactive GPG 2.x scripted-keygen
	// incantation (verified directly against the real gpg 2.4.4 binary on this box before this
	// code was written, not assumed from documentation).
	genCmd := exec.Command("gpg", "--homedir", homeDir, "--batch", "--passphrase", "",
		"--pinentry-mode", "loopback", "--quick-gen-key", uid, "rsa4096", "sign", "0")
	var genErrBuf bytes.Buffer
	genCmd.Stderr = &genErrBuf
	if runErr := genCmd.Run(); runErr != nil {
		return "", "", false, fmt.Errorf("gpgkey: generate key: %w: %s", runErr, genErrBuf.String())
	}

	id, findErr = findKeyID(homeDir, uid)
	if findErr != nil {
		return "", "", false, findErr
	}
	if id == "" {
		return "", "", false, fmt.Errorf("gpgkey: key generation reported success but no key was found for %q", uid)
	}
	armored, expErr := exportPublic(homeDir, id)
	if expErr != nil {
		return "", "", false, expErr
	}
	return armored, id, true, nil
}

// findKeyID looks for an existing secret key matching uid in homeDir's own keyring, parsing
// `--with-colons` machine-readable output (the real, documented, stable GnuPG interface for
// scripting -- never parse gpg's human-readable output). Returns "" (not an error) if none found
// yet -- a real, honest, expected state on first run, same convention sshkey.LoadOrGenerate's own
// os.Stat-not-exists branch already follows.
func findKeyID(homeDir, uid string) (string, error) {
	cmd := exec.Command("gpg", "--homedir", homeDir, "--batch", "--list-secret-keys", "--with-colons", uid)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	if runErr != nil {
		// A non-zero exit here means "no matching key" (gpg's own real, standard behavior for
		// an unmatched search term), not a real error -- same "not found isn't an error" shape
		// os.Stat(...os.IsNotExist) already gets from the OS for the SSH key path.
		if strings.Contains(errBuf.String(), "No secret key") || out.Len() == 0 {
			return "", nil
		}
	}
	for _, line := range strings.Split(out.String(), "\n") {
		fields := strings.Split(line, ":")
		if len(fields) > 4 && fields[0] == "sec" {
			return fields[4], nil // field 4 (0-indexed) is the long key ID in --with-colons output
		}
	}
	return "", nil
}

// exportPublic returns the ASCII-armored public key block for keyID -- --export (not
// --export-secret-keys) so the private key material is structurally impossible to return from
// this function, not just a convention this package promises to follow.
func exportPublic(homeDir, keyID string) (string, error) {
	cmd := exec.Command("gpg", "--homedir", homeDir, "--batch", "--armor", "--export", keyID)
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gpgkey: export public key %s: %w: %s", keyID, err, errBuf.String())
	}
	armored := strings.TrimRight(out.String(), "\n")
	if armored == "" {
		return "", fmt.Errorf("gpgkey: export public key %s: empty output", keyID)
	}
	return armored, nil
}
