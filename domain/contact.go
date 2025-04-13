package domain

import (
	"time"

	"github.com/arrno/chanman/util"
)

// ContactRequest represents a contact form submission
type ContactRequest struct {
	Id        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Message   string    `db:"message" json:"message"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// NewContactRequest creates a new contact request
func NewContactRequest(name, email, message string) *ContactRequest {
	return &ContactRequest{
		Id:        util.Nanoid(),
		Name:      name,
		Email:     email,
		Message:   message,
		CreatedAt: time.Now(),
	}
}
