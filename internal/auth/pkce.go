package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// unreserved URI characters per RFC 7636
const verifierChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

// GenerateCodeVerifier generates a cryptographically random code verifier (43-128 chars).
func GenerateCodeVerifier() (string, error) {
	const length = 64
	n := len(verifierChars)
	// Rejection threshold to avoid modulo bias: discard bytes >= maxValid
	maxValid := byte(256 - 256%n)

	out := make([]byte, length)
	buf := make([]byte, 1)
	for i := 0; i < length; {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		if buf[0] >= maxValid {
			continue // reject to avoid bias
		}
		out[i] = verifierChars[buf[0]%byte(n)]
		i++
	}
	return string(out), nil
}

// CodeChallengeS256 computes base64url(SHA256(verifier)) per RFC 7636.
func CodeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// GenerateState generates a cryptographically random state parameter (32 bytes, base64url).
func GenerateState() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
