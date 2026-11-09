package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	gossh "golang.org/x/crypto/ssh"
)

// Identity is a visitor identity derived from SSH credentials or anonymous access.
type Identity struct {
	fingerprint string
	nickname    string
	account     string
	anonymous   bool
}

// Anonymous returns a non-persistent identity.
func Anonymous() Identity {
	return Identity{anonymous: true, fingerprint: "anonymous"}
}

// ForFingerprint constructs an identity for store lookups by stable fingerprint.
func ForFingerprint(fp string) Identity {
	return Identity{fingerprint: fp}
}

// FromPublicKey derives a stable fingerprint identity from an SSH public key.
func FromPublicKey(key gossh.PublicKey) Identity {
	if key == nil {
		return Anonymous()
	}
	sum := sha256.Sum256(key.Marshal())
	return Identity{fingerprint: hex.EncodeToString(sum[:])}
}

// FromAccount derives a stable persistence key for password-authenticated users.
func FromAccount(username string) Identity {
	sum := sha256.Sum256([]byte("account:" + username))
	return Identity{account: username, fingerprint: hex.EncodeToString(sum[:])}
}

// Fingerprint returns the stable key for persistence (empty when anonymous).
func (i Identity) Fingerprint() string {
	if i.anonymous {
		return ""
	}
	return i.fingerprint
}

// Account returns the SSH username when password auth established the session.
func (i Identity) Account() string { return i.account }

// Nickname returns the reserved display name if any.
func (i Identity) Nickname() string { return i.nickname }

// WithNickname returns a copy with nickname set.
func (i Identity) WithNickname(name string) Identity {
	i.nickname = name
	return i
}

// Anonymous reports whether this session has no persistent identity.
func (i Identity) Anonymous() bool { return i.anonymous }

// String is a short display label for presence lists.
func (i Identity) String() string {
	if i.nickname != "" {
		return i.nickname
	}
	if i.account != "" {
		return i.account
	}
	if i.anonymous {
		return "guest"
	}
	if len(i.fingerprint) >= 8 {
		return "key-" + i.fingerprint[:8]
	}
	return "visitor"
}

// ValidateNickname checks reservation rules.
func ValidateNickname(name string) error {
	if name == "" {
		return fmt.Errorf("portico: nickname must not be empty")
	}
	if len(name) > 32 {
		return fmt.Errorf("portico: nickname too long")
	}
	return nil
}

// ValidateUsername checks account names used as store keys and filenames.
func ValidateUsername(name string) error {
	if name == "" || len(name) > 32 {
		return fmt.Errorf("portico: invalid username")
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return fmt.Errorf("portico: invalid username")
		}
	}
	return nil
}
