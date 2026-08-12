package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

func NewPKCE() (verifier, challenge string, err error) {
	verifier, err = randomURLSafeString(64)
	if err != nil {
		return "", "", err
	}
	challenge = computeCodeChallenge(verifier)
	return verifier, challenge, nil
}

func randomURLSafeString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func computeCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
