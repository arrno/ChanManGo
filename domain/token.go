package domain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Token represents an API token with its raw value, hashed value, and optional salt
type Token struct {
	key       []byte `json:"-"`
	Raw       string `json:"raw"`
	Encrypted string `json:"encrypted"`
}

func NewToken(raw, localSecret string) (Token, error) {
	t := Token{
		Raw: raw,
		key: []byte(localSecret),
	}
	t.Encrypt()
	return t, nil
}

func (t Token) Validate(against string) bool {
	err := t.Decrypt()
	if err != nil {
		return false
	}
	return t.Raw == against
}

func (t *Token) Encrypt() error {
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

func GenerateRandomSecretKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	return key, nil
}
