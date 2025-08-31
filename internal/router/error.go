package router

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Status  int    `json:"status"`
	Path    string `json:"path"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func (e *ErrorResponse) WriteTo(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	if e.Error != "" {
		e.Message = e.Error
	}
	response, _ := json.Marshal(e)
	_, err := w.Write(response)
	if err != nil {
		slog.Error("Failed to write error response", "error", err)
		return
	}
}
