package router

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Status  int    `json:"status"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (e *ErrorResponse) WriteTo(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	if e.Error != "" {
		e.Message = e.Error
	}
	response, _ := json.Marshal(e)
	_, err := w.Write(response)
	if err != nil {
		return err
	}
	return nil
}
