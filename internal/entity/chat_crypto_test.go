package entity

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestVerifyDiscordSignatureRoundTrip(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	pubHex := hex.EncodeToString(pub)
	timestamp := "1700000000"
	body := []byte(`{"type":1}`)
	sig := hex.EncodeToString(ed25519.Sign(priv, append([]byte(timestamp), body...)))

	if !VerifyDiscordSignature(pubHex, timestamp, sig, body) {
		t.Fatal("a valid Discord signature must verify")
	}
	if VerifyDiscordSignature(pubHex, timestamp, sig, []byte(`{"type":2}`)) {
		t.Fatal("a tampered body must not verify")
	}
	if VerifyDiscordSignature(pubHex, "1700000001", sig, body) {
		t.Fatal("a tampered timestamp must not verify")
	}
	if VerifyDiscordSignature("", timestamp, sig, body) {
		t.Fatal("an empty public key must not verify")
	}
	other, _, _ := ed25519.GenerateKey(nil)
	if VerifyDiscordSignature(hex.EncodeToString(other), timestamp, sig, body) {
		t.Fatal("a different key must not verify")
	}
}

func TestVerifySlackSignatureRoundTrip(t *testing.T) {
	secret := "s3cr3t"
	now := time.Unix(1700000000, 0)
	timestamp := "1700000000"
	body := []byte("payload=%7B%22type%22%3A%22block_actions%22%7D")

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte("v0:" + timestamp + ":"))
	mac.Write(body)
	sig := "v0=" + hex.EncodeToString(mac.Sum(nil))

	if !VerifySlackSignature(secret, timestamp, sig, body, now, 5*time.Minute) {
		t.Fatal("a valid Slack signature must verify")
	}
	if VerifySlackSignature(secret, timestamp, sig, []byte("payload=tampered"), now, 5*time.Minute) {
		t.Fatal("a tampered body must not verify")
	}
	if VerifySlackSignature("wrong", timestamp, sig, body, now, 5*time.Minute) {
		t.Fatal("a wrong secret must not verify")
	}
	stale := now.Add(10 * time.Minute)
	if VerifySlackSignature(secret, timestamp, sig, body, stale, 5*time.Minute) {
		t.Fatal("a request outside the skew window must not verify")
	}
}
