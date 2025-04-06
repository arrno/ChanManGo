package domain

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/arrno/chanman/util"
)

var (
	ErrTokenInvalid = errors.New("Token invalid")
	ErrTokenExpired = errors.New("Token expired")
)

// Token represents an API token with its raw value, hashed value, and optional salt
type Token struct {
	// key       []byte     `json:"-"`
	Raw       string     `json:"raw"`
	Salt      string     `json:"salt"`
	Hashed    string     `json:"hash"`
	Identity  string     `json:"identity"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
}

func NewToken(raw, identity string) *Token {
	t := &Token{
		Raw:       raw,
		Identity:  identity,
		CreatedAt: time.Now(),
	}
	return t
}

// TODO we need to honor this in the client_secret_tokens table
func NewTokenWithExpiry(raw, identity string, expiryDuration time.Duration) (Token, error) {
	expires := time.Now().Add(expiryDuration)
	t := Token{
		Raw:       raw,
		Identity:  identity,
		ExpiresAt: &expires,
		CreatedAt: time.Now(),
	}
	t.Hash()
	return t, nil
}

func (t *Token) Hash() *Token {
	salt, err := util.GenerateSalt(16) // 16 bytes is a good salt size
	if err != nil {
		panic(err)
	}
	hasher := sha256.New()
	hasher.Write([]byte(t.Raw + salt))
	hashBytes := hasher.Sum(nil)
	t.Salt = salt
	t.Hashed = hex.EncodeToString(hashBytes)
	return t
}

func (t Token) Validate(against string) error {
	// Check if token is expired
	if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
		return ErrTokenExpired
	}

	hasher := sha256.New()
	hasher.Write([]byte(against + t.Salt))
	hashBytes := hasher.Sum(nil)
	againstHash := hex.EncodeToString(hashBytes)

	// Use a constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(t.Hashed), []byte(againstHash)) == 1 {
		return nil
	}
	return ErrTokenInvalid
}

func (t Token) Serialize() string {
	return fmt.Sprintf("%s::%s", t.Identity, t.Raw)
}

func (t *Token) Deserialize(str string) bool {
	parts := strings.Split(str, "::")
	if len(parts) < 2 {
		return false
	}
	t.Raw = parts[len(parts)-1]
	t.Identity = strings.Join(parts[0:len(parts)-1], "::")
	return true
}

func TokenFromSecret(s ClientSecret) *Token {
	t := new(Token)
	t.Salt = s.Salt
	t.Hashed = s.EncryptedSecret
	t.Identity = s.ClientId
	return t
}

// func (t Token) Validate(against string) bool {
// 	// Check if token is expired
// 	if t.ExpiresAt != nil && time.Now().After(*t.ExpiresAt) {
// 		return false
// 	}

// 	// Decrypt and validate
// 	err := t.Decrypt()
// 	if err != nil {
// 		return false
// 	}

// 	// Use a constant-time comparison to prevent timing attacks
// 	return subtle.ConstantTimeCompare([]byte(t.Raw), []byte(against)) == 1
// }

// func (t *Token) Encrypt() error {
// 	block, err := aes.NewCipher(t.key)
// 	if err != nil {
// 		return err
// 	}

// 	aesGCM, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return err
// 	}

// 	nonce := make([]byte, aesGCM.NonceSize())
// 	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
// 		return err
// 	}

// 	ciphertext := aesGCM.Seal(nonce, nonce, []byte(t.Raw), nil)
// 	t.Encrypted = base64.StdEncoding.EncodeToString(ciphertext)
// 	return nil
// }

// func (t *Token) Decrypt() error {
// 	ciphertext, err := base64.StdEncoding.DecodeString(t.Encrypted)
// 	if err != nil {
// 		return err
// 	}

// 	block, err := aes.NewCipher(t.key)
// 	if err != nil {
// 		return err
// 	}

// 	aesGCM, err := cipher.NewGCM(block)
// 	if err != nil {
// 		return err
// 	}

// 	nonceSize := aesGCM.NonceSize()
// 	if len(ciphertext) < nonceSize {
// 		return fmt.Errorf("ciphertext too short")
// 	}

// 	nonce, encryptedData := ciphertext[:nonceSize], ciphertext[nonceSize:]
// 	plaintext, err := aesGCM.Open(nil, nonce, encryptedData, nil)
// 	if err != nil {
// 		return err
// 	}
// 	t.Raw = string(plaintext)
// 	return nil
// }
