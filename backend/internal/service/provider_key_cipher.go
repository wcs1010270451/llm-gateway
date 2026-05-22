package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const providerKeyCipherPrefix = "enc:v1:"

type ProviderKeyCipher struct {
	key []byte
}

func NewProviderKeyCipher(secret string) *ProviderKeyCipher {
	sum := sha256.Sum256([]byte(secret))
	return &ProviderKeyCipher{key: sum[:]}
}

func (c *ProviderKeyCipher) Encrypt(plainText string) (string, error) {
	plainText = strings.TrimSpace(plainText)
	if plainText == "" {
		return "", nil
	}
	if strings.HasPrefix(plainText, providerKeyCipherPrefix) {
		return plainText, nil
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := gcm.Seal(nil, nonce, []byte(plainText), nil)
	payload := append(nonce, cipherText...)
	return providerKeyCipherPrefix + base64.RawStdEncoding.EncodeToString(payload), nil
}

func (c *ProviderKeyCipher) Decrypt(storedValue string) (string, error) {
	storedValue = strings.TrimSpace(storedValue)
	if storedValue == "" {
		return "", nil
	}
	if !strings.HasPrefix(storedValue, providerKeyCipherPrefix) {
		return storedValue, nil
	}

	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(storedValue, providerKeyCipherPrefix))
	if err != nil {
		return "", fmt.Errorf("decode provider api key: %w", err)
	}

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", fmt.Errorf("provider api key ciphertext is too short")
	}

	nonce := payload[:gcm.NonceSize()]
	cipherText := payload[gcm.NonceSize():]
	plainText, err := gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt provider api key: %w", err)
	}
	return string(plainText), nil
}
