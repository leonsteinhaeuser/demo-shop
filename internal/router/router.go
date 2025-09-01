package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ApiMeta defines the interface for API route objects
// that can be used in a router.
// It provides methods to retrieve the API version and kind of the object.
// This interface is used to ensure that all API route objects
// implement the necessary methods to be recognized by the router.
// It is typically implemented by structs that represent API resources
// such as schemas, users, or other entities in the API.
// It is used to provide a consistent way to access the version and kind
// of the API objects, which is essential for routing and handling requests
// in a RESTful API framework.
type ApiMeta interface {
	// GetApiVersion returns the api object version
	GetApiVersion() string
	// GetGroup returns the api object group
	GetGroup() string
	// GetKind returns the api object kind
	GetKind() string
}

type PathObject struct {
	Path   string
	Method string
	Func   http.HandlerFunc
}

type ApiSpec interface {
	Routes() []PathObject
}

type ApiObject interface {
	ApiMeta
	ApiSpec
}

var (
	ErrUnableToRegisterAlreadyExists = fmt.Errorf("unable to register: object already exists")
	ErrObjectStorageNotImplemented   = fmt.Errorf("object storage interface not implemented")

	DefaultRouter = &Router{
		apiObjects: make(map[string]ApiObject),
		apiSpec:    []ApiObjectMeta{},
		readyCh:    make(chan bool, 1),
		livenessCh: make(chan bool, 2),
	}
)

type Router struct {
	// apiObjects is a map of API route objects
	// The key is the full path of the API object
	apiObjects map[string]ApiObject
	// rawRoutes is a map of HTTP method and path to handler functions
	rawRoutes map[string]http.HandlerFunc

	// apiSpec is a map of API paths and their methods
	apiSpec []ApiObjectMeta

	readyCh    chan bool
	livenessCh chan bool
}

type ApiObjectMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (r *Router) SetReady(prop bool) {
	go func() {
		r.readyCh <- prop
	}()
}

func (r *Router) SetLiveness(prop bool) {
	go func() {
		r.livenessCh <- prop
	}()
}

// RegisterPath registers a raw HTTP route with the router.
// It allows you to register a specific HTTP method and path with a handler function.
// This is useful for defining custom routes that do not correspond to a specific API object.
func (r *Router) RegisterPath(method, path string, handler http.HandlerFunc) error {
	if r.rawRoutes == nil {
		r.rawRoutes = make(map[string]http.HandlerFunc)
	}
	routePath := fmt.Sprintf("%s %s", method, path)
	// check if the method and path combination already exists
	if _, exists := r.rawRoutes[routePath]; exists {
		// if it exists, return an error
		return fmt.Errorf("route already exists: %s", routePath)
	}
	r.rawRoutes[routePath] = handler
	return nil
}

// Register registers an API route object with the router.
func (r *Router) Register(obj ApiObject) error {
	// define the fully qualified path (fqp) for the API object
	fqp := path.Join("/api", obj.GetApiVersion(), obj.GetGroup(), obj.GetKind())
	// look for existing objects with the same fqp
	if _, ok := r.apiObjects[fqp]; ok {
		// if an object with the same fqp already exists, return an error
		return fmt.Errorf("%w: %s", ErrUnableToRegisterAlreadyExists, fqp)
	}
	// if no existing object is found, register the new object
	r.apiObjects[fqp] = obj
	return nil
}

func (r *Router) Build(mux *http.ServeMux) error {
	// iterate over all registered API objects
	for fqp, obj := range r.apiObjects {
		slog.Info("Registering API object", "path", fqp)
		for _, pobj := range obj.Routes() {
			fpath := path.Join(fqp, pobj.Path)
			slog.Info("Route", "method", pobj.Method, "path", fpath)
			r.apiSpec = append(r.apiSpec, ApiObjectMeta{
				Path:   fpath,
				Method: pobj.Method,
			})
			mux.HandleFunc(pobj.Method+" "+fpath, pobj.Func)
		}
	}
	r.apiSpec = append(r.apiSpec,
		ApiObjectMeta{
			Path:   "/metrics",
			Method: "GET",
		}, ApiObjectMeta{
			Path:   "/health/liveness",
			Method: "GET",
		}, ApiObjectMeta{
			Path:   "/health/readiness",
			Method: "GET",
		},
	)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("GET /health/readiness", httpHealthz(r.readyCh))
	mux.HandleFunc("GET /health/liveness", httpHealthz(r.livenessCh))
	mux.HandleFunc("/api/metadata", func(wrt http.ResponseWriter, req *http.Request) {
		wrt.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(wrt).Encode(r.apiSpec); err != nil {
			(&ErrorResponse{
				Status:  http.StatusInternalServerError,
				Path:    req.URL.Path,
				Message: "Failed to encode API metadata",
				Error:   err.Error(),
			}).WriteTo(wrt)
			return
		}
	})
	return nil
}

func EnableCorsHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8088")
		// Allow credentials for session management
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Allow specific HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// Allow specific headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, Accept, Origin")

		// Set max age for preflight cache
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Add no-cache headers for development
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func httpHealthz(readyCh <-chan bool) func(w http.ResponseWriter, r *http.Request) {
	isReady := false
	go func() {
		for {
			isReady = <-readyCh
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		if !isReady {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}
}
