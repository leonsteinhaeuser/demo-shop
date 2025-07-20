package router

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"
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
}

type ApiObjectMeta struct {
	Path   string `json:"path"`
	Method string `json:"method"`
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
