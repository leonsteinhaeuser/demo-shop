package handlers

import (
	"net/http"
	"strconv"
)

// QueryStringValue retrieves a string value from the query parameters.
func QueryStringValue(r *http.Request, key string) string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return ""
	}
	return value
}

// QueryIntValue retrieves an integer value from the query parameters.
func QueryIntValue(r *http.Request, key string) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return 0, nil
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return intValue, nil
}

// QueryBoolValue retrieves a boolean value from the query parameters.
func QueryBoolValue(r *http.Request, key string) (bool, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return false, nil
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, err
	}
	return boolValue, nil
}
