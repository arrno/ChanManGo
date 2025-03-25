package domain

import (
	"golang.org/x/crypto/bcrypt"
)

// Token represents an API token with its raw value, hashed value, and optional salt
type Token struct {
	Raw  string `json:"raw"`
	Hash string `json:"hash"`
}

func NewToken(raw string) (Token, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return Token{}, err
	}

	return Token{
		Raw:  raw,
		Hash: string(hash),
	}, nil
}

func (t Token) Validate() bool {
	return nil == bcrypt.CompareHashAndPassword([]byte(t.Hash), []byte(t.Raw))
}
