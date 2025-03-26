package controller

import (
	"encoding/json"
	"io"
	"net/http"
)

func UnmarshalTo[T any](target *T, r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, target)
	if err != nil {
		return err
	}
	return nil
}
