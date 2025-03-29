package domain

import (
	"net/http"
	"time"

	"github.com/arrno/chanman/util"
)

type ClientPermission struct {
	Id         string     `json:"id"`
	ClientId   string     `json:"clientId"`
	Read       bool       `json:"read"`
	Write      bool       `json:"write"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ArchivedAt *time.Time `json:"archivedAt"`
}

func NewReadWrite(clientId string) *ClientPermission {
	now := time.Now()
	return &ClientPermission{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		Read:      true,
		Write:     true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewReadOnly(clientId string) *ClientPermission {
	now := time.Now()
	return &ClientPermission{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		Read:      true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (cp *ClientPermission) Enforce(with ClientPermission) error {
	if with.Read && !cp.Read {
		return ErrInsufficientPermissions
	}
	if with.Write && !cp.Write {
		return ErrInsufficientPermissions
	}
	return nil
}

func (cp *ClientPermission) AddToHeaders(r *http.Request) {
	if cp.Read {
		r.Header.Add("read", "true")
	}
	if cp.Write {
		r.Header.Add("write", "true")
	}
}

func (cp *ClientPermission) Archive() {
	now := time.Now()
	cp.ArchivedAt = &now
	cp.UpdatedAt = now
}

func (cp *ClientPermission) UpdatePermissions(read, write bool) {
	cp.Read = read
	cp.Write = write
	cp.UpdatedAt = time.Now()
}
