package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// GetUUIDFromPathValue retrieves a UUID from the request path value.
func GetUUIDFromPathValue(r *http.Request, name string) (uuid.UUID, error) {
	sid := r.PathValue(name)
	if sid == "" {
		return uuid.Nil, fmt.Errorf("missing path value for property: %s", name)
	}
	id, err := uuid.Parse(sid)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}
