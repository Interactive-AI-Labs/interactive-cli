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
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i := range buf {
		out[i] = verifierChars[int(buf[i])%len(verifierChars)]
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
