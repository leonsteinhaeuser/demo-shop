package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/leonsteinhaeuser/demo-shop/internal/router"
)

// HttpPost handles HTTP POST requests.
func HttpPost[T any](storeFunc func(context.Context, *http.Request, *T) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		obj := new(T)
		err := json.NewDecoder(r.Body).Decode(obj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Invalid request body",
				Error:   err.Error(),
			}).WriteTo(w)
		}

		err = storeFunc(ctx, r, obj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Path:    r.URL.Path,
				Message: "Failed to store resource",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Path:    r.URL.Path,
				Message: "Failed to encode response",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
	}

}

type FilterObjectList struct {
	Limit int
	Page  int
}

func HttpList[T any](fetchFunc func(context.Context, *http.Request, FilterObjectList) ([]T, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		limit, err := QueryIntValue(r, "limit")
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Invalid limit query parameter",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}

		page, err := QueryIntValue(r, "page")
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Invalid page query parameter",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}

		fobj := FilterObjectList{
			Limit: limit,
			Page:  page,
		}

		result, err := fetchFunc(ctx, r, fobj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Failed to fetch resources",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Path:    r.URL.Path,
				Message: "Failed to encode response",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
	}
}

func HttpGet[T any](fetchFunc func(context.Context, *http.Request) (*T, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		result, err := fetchFunc(ctx, r)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusNotFound,
				Path:    r.URL.Path,
				Message: "Resource not found",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Failed to encode response",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
	}
}

func HttpUpdate[T any](updateFunc func(context.Context, *http.Request, *T) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		obj := new(T)
		err := json.NewDecoder(r.Body).Decode(obj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Invalid request body",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
		err = updateFunc(ctx, r, obj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Failed to update resource",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		if err := json.NewEncoder(w).Encode(obj); err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Path:    r.URL.Path,
				Message: "Failed to encode response",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
	}
}

func HttpDelete[T any](deleteFunc func(context.Context, *http.Request, *T) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		obj := new(T)
		err := json.NewDecoder(r.Body).Decode(obj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Invalid request body",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
		err = deleteFunc(ctx, r, obj)
		if err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusBadRequest,
				Path:    r.URL.Path,
				Message: "Failed to delete resource",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		if err := json.NewEncoder(w).Encode(nil); err != nil {
			(&router.ErrorResponse{
				Status:  http.StatusInternalServerError,
				Path:    r.URL.Path,
				Message: "Failed to encode response",
				Error:   err.Error(),
			}).WriteTo(w)
			return
		}
	}
}

func HttpHealthz(readyCh <-chan bool) func(w http.ResponseWriter, r *http.Request) {
	isReady := false
	go func() {
		for {
			ready := <-readyCh
			isReady = ready
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		if !isReady {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
