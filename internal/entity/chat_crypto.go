package entity

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func VerifySlackSignature(signingSecret, timestamp, signature string, body []byte, now time.Time, skew time.Duration) bool {
	if signingSecret == "" || timestamp == "" || signature == "" {
		return false
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false
	}
	if diff := now.Sub(time.Unix(ts, 0)); diff > skew || diff < -skew {
		return false
	}
	mac := hmac.New(sha256.New, []byte(signingSecret))
	mac.Write([]byte("v0:" + timestamp + ":"))
	mac.Write(body)
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func VerifyDiscordSignature(publicKeyHex, timestamp, signatureHex string, body []byte) bool {
	if publicKeyHex == "" || timestamp == "" || signatureHex == "" {
		return false
	}
	key, err := hex.DecodeString(publicKeyHex)
	if err != nil || len(key) != ed25519.PublicKeySize {
		return false
	}
	sig, err := hex.DecodeString(signatureHex)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return false
	}
	msg := make([]byte, 0, len(timestamp)+len(body))
	msg = append(msg, timestamp...)
	msg = append(msg, body...)
	return ed25519.Verify(ed25519.PublicKey(key), msg, sig)
}

func SecretEqual(a, b string) bool {
	return hmac.Equal([]byte(a), []byte(b))
}

func TrimBearer(header string) string {
	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
}
