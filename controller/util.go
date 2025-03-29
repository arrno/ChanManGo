package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

// RespondWithError sends a standardized error response
func RespondWithError(w http.ResponseWriter, code int, message string, err error) {
	response := StandardResponse{
		Success: false,
		Message: message,
	}

	// Only include detailed error in development, log it either way
	if err != nil {
		log.Error().Err(err).Msg(message)
		// In production, don't expose internal errors
		// response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithJSON sends a standardized success response
func RespondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response := StandardResponse{
		Success: true,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

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
