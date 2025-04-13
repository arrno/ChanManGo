package domain

import (
	"net/http"
	"time"

	"github.com/arrno/chanman/util"
)

type Tier string

const (
	FreeTier Tier = "free"
	ProTier  Tier = "pro"
)

type ClientPermission struct {
	Id         string     `db:"id" json:"id"`
	ClientId   string     `db:"client_id" json:"clientId"`
	Read       bool       `db:"allow_read" json:"read"`
	Write      bool       `db:"allow_write" json:"write"`
	Create     bool       `db:"allow_create" json:"create"`
	Tier       Tier       `db:"tier" json:"tier"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	ArchivedAt *time.Time `db:"archived_at" json:"archivedAt"`
}

func NewReadWrite(clientId string) *ClientPermission {
	now := time.Now()
	return &ClientPermission{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		Read:      true,
		Write:     true,
		Tier:      FreeTier,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewAll(clientId string) *ClientPermission {
	now := time.Now()
	return &ClientPermission{
		Id:        util.Nanoid(),
		ClientId:  clientId,
		Read:      true,
		Write:     true,
		Create:    true,
		Tier:      ProTier,
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
		Tier:      FreeTier,
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
	if with.Create && !cp.Create {
		return ErrInsufficientPermissions
	}
	if with.Tier == ProTier && cp.Tier != ProTier {
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
	if cp.Create {
		r.Header.Add("create", "true")
	}
	r.Header.Add("tier", string(cp.Tier))
}

func (cp *ClientPermission) Archive() {
	now := time.Now()
	cp.ArchivedAt = &now
	cp.UpdatedAt = now
}

func (cp *ClientPermission) UpdatePermissions(read, write, create bool, tier Tier) {
	cp.Read = read
	cp.Write = write
	cp.Create = create
	cp.Tier = tier
	cp.UpdatedAt = time.Now()
}
