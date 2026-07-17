package secretbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/opsybot/opsybot/internal/config"
)

const (
	keyVersion = 0x01
	nonceSize  = 12
	infoLabel  = "opsybot/secretbox"
)

var (
	ErrDisabled   = errors.New("secretbox disabled: no auth.secret_key configured")
	ErrDecrypt    = errors.New("secretbox decrypt failed")
	ErrCiphertext = errors.New("secretbox malformed ciphertext")
)

type Client struct {
	current  cipher.AEAD
	previous cipher.AEAD
}

func New(cfg config.Auth) (Client, error) {
	c := Client{}
	if cfg.SecretKey != "" {
		aead, err := aeadFromSecret(cfg.SecretKey)
		if err != nil {
			return Client{}, err
		}
		c.current = aead
	}
	if cfg.SecretKeyPrevious != "" {
		aead, err := aeadFromSecret(cfg.SecretKeyPrevious)
		if err != nil {
			return Client{}, err
		}
		c.previous = aead
	}
	return c, nil
}

func aeadFromSecret(secret string) (cipher.AEAD, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(hkdf.New(sha256.New, []byte(secret), nil, []byte(infoLabel)), key); err != nil {
		return nil, fmt.Errorf("derive key: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("new cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("new gcm: %w", err)
	}
	return aead, nil
}

func (c Client) Enabled() bool { return c.current != nil }

func (c Client) Encrypt(plaintext []byte) ([]byte, error) {
	if c.current == nil {
		return nil, ErrDisabled
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}
	sealed := c.current.Seal(nil, nonce, plaintext, nil)
	out := make([]byte, 0, 1+nonceSize+len(sealed))
	out = append(out, keyVersion)
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

func (c Client) Decrypt(data []byte) ([]byte, error) {
	if c.current == nil {
		return nil, ErrDisabled
	}
	if len(data) < 1+nonceSize+1 || data[0] != keyVersion {
		return nil, ErrCiphertext
	}
	nonce := data[1 : 1+nonceSize]
	ciphertext := data[1+nonceSize:]
	if plaintext, err := c.current.Open(nil, nonce, ciphertext, nil); err == nil {
		return plaintext, nil
	}
	if c.previous != nil {
		if plaintext, err := c.previous.Open(nil, nonce, ciphertext, nil); err == nil {
			return plaintext, nil
		}
	}
	return nil, ErrDecrypt
}
