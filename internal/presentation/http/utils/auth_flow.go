package presentationhttputils

import (
	"crypto/sha256"
	"encoding/hex"
)

// GoogleBindingCookieName correlates independent browser flows without trusting state.
func GoogleBindingCookieName(state string) string {
	digest := sha256.Sum256([]byte(state))
	return "__Host-lockmate-google-" + hex.EncodeToString(digest[:16])
}
