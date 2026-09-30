package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

func ValidHash(value string) bool {
	raw, err := hex.DecodeString(value)
	return err == nil && len(raw) == sha256.Size && value == strings.ToLower(value)
}
func Matches(credential, expected string) bool {
	if credential == "" {
		return false
	}
	sum := sha256.Sum256([]byte(credential))
	return subtle.ConstantTimeCompare([]byte(hex.EncodeToString(sum[:])), []byte(expected)) == 1
}
