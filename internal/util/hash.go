package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashContent returns a truncated SHA256 hash of the content
func HashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])[:16]
}
