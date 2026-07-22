package entity

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"golang.org/x/crypto/argon2"
)

var ErrPasswordMismatch = errors.New("password mismatch")

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("random bytes: %w", err)
	}
	return b, nil
}

func RandomSlugSuffix() (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(SlugSuffixSpan))
	if err != nil {
		return 0, fmt.Errorf("random slug suffix: %w", err)
	}
	return SlugSuffixMin + int(n.Int64()), nil
}

func GenerateToken(byteLen int) (string, error) {
	b, err := randomBytes(byteLen)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GenerateHexToken(byteLen int) (string, error) {
	b, err := randomBytes(byteLen)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func SignBody(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return SourceSignaturePrefix + hex.EncodeToString(mac.Sum(nil))
}

func VerifyBodySignature(secrets []string, body []byte, signature string) bool {
	provided := strings.TrimSpace(signature)
	if provided == "" {
		return false
	}
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(SignBody(secret, body)), []byte(provided)) == 1 {
			return true
		}
	}
	return false
}

func HashPassword(password string) (string, error) {
	salt, err := randomBytes(ArgonSaltLength)
	if err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, ArgonTime, ArgonMemoryKiB, ArgonThreads, ArgonKeyLength)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, ArgonMemoryKiB, ArgonTime, ArgonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

func VerifyPassword(hash, password string) error {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return ErrPasswordMismatch
	}
	var version, memory, timeCost, threads int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return ErrPasswordMismatch
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return ErrPasswordMismatch
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return ErrPasswordMismatch
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return ErrPasswordMismatch
	}
	got := argon2.IDKey([]byte(password), salt, uint32(timeCost), uint32(memory), uint8(threads), uint32(len(want)))
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return ErrPasswordMismatch
	}
	return nil
}
