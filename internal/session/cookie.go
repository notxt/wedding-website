package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"strings"
)

type Signer struct {
	secret []byte
}

func New(secret string) *Signer {
	return &Signer{secret: []byte(secret)}
}

func (s *Signer) Sign(marker string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(marker))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return marker + "." + sig
}

func (s *Signer) Verify(value string) bool {
	i := strings.LastIndex(value, ".")
	if i < 0 {
		return false
	}
	expected := s.Sign(value[:i])
	return subtle.ConstantTimeCompare([]byte(value), []byte(expected)) == 1
}
