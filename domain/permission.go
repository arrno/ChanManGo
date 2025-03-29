package domain

import (
	"net/http"

	"github.com/arrno/chanman/util"
)

type ClientPermission struct {
	Id       string `json:"id"`
	ClientId string `json:"clientId"`
	Read     bool   `json:"read"`
	Write    bool   `json:"write"`
}

func NewReadWrite(clientId string) *ClientPermission {
	return &ClientPermission{
		Id:       util.Nanoid(),
		ClientId: clientId,
		Read:     true,
		Write:    true,
	}
}

func NewReadOnly(clientId string) *ClientPermission {
	return &ClientPermission{
		Id:       util.Nanoid(),
		ClientId: clientId,
		Read:     true,
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
