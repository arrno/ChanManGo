package domain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// Token represents an API token with its raw value, hashed value, and optional salt
type Token struct {
	key       []byte     `json:"-"`
	Raw       string     `json:"raw"`
	Encrypted string     `json:"encrypted"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

func NewToken(raw, localSecret string) (Token, error) {
	t := Token{
		Raw:       raw,
		key:       []byte(localSecret),
		CreatedAt: time.Now(),
	}
	t.Encrypt()
	return t, nil
}

func NewTokenWithExpiry(raw, localSecret string, expiryDuration time.Duration) (Token, error) {
	expires := time.Now().Add(expiryDuration)
	t := Token{
		Raw:       raw,
		key:       []byte(localSecret),
		ExpiresAt: &expires,
		CreatedAt: time.Now(),
	}
	t.Encrypt()
	return t, nil
}

func (t Token) Validate(against string) bool {
	// Check if token is expired
	if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
		return false
	}

	// Decrypt and validate
	err := t.Decrypt()
	if err != nil {
		return false
	}

	// Use a constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(t.Raw), []byte(against)) == 1
}

func (t *Token) Encrypt() {
	t.Hash() // For now...
}

func (t *Token) Hash() {
	hasher := sha256.New()
	hasher.Write([]byte(t.Raw))
	hashBytes := hasher.Sum(nil)
	t.Encrypted = hex.EncodeToString(hashBytes)
}

func (t *Token) encrypt() error {
	block, err := aes.NewCipher(t.key)
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(t.Raw), nil)
	t.Encrypted = base64.StdEncoding.EncodeToString(ciphertext)
	return nil
}

func (t *Token) Decrypt() error {
	ciphertext, err := base64.StdEncoding.DecodeString(t.Encrypted)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(t.key)
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return err
	}
	t.Raw = string(plaintext)
	return nil
}
