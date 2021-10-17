package chi

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/tal-tech/go-zero/rest/httpx"
	"github.com/tal-tech/go-zero/rest/pathvar"
)

var (
	// ErrInvalidMethod is an error that indicates not a valid http method.
	ErrInvalidMethod = errors.New("not a valid http method")
	// ErrInvalidPath is an error that indicates path is not start with /.
	ErrInvalidPath = errors.New("path must begin with '/'")
)

type chiRouter struct {
	mux *chi.Mux
}

// NewRouter returns a chi.Router.
func NewRouter() httpx.Router {
	return &chiRouter{
		mux: chi.NewRouter(),
	}
}

func (pr *chiRouter) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) {
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' {
		return ErrInvalidPath
	}

	pr.getHandleFunc(method)(reqPath, func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, pr.withUrlParamsToContext(r))
	})
	return nil
}

func (pr *chiRouter) withUrlParamsToContext(r *http.Request) *http.Request {
	if ctx := chi.RouteContext(r.Context()); ctx != nil {
		urlParams := ctx.URLParams
		params := make(map[string]string)
		for i := 0; i < len(urlParams.Values); i++ {
			params[urlParams.Keys[i]] = urlParams.Values[i]
		}
		if len(params) > 0 {
			r = pathvar.WithVars(r, params)
		}
	}
	return r
}

func (pr *chiRouter) getHandleFunc(method string) func(pattern string, handlerFn http.HandlerFunc) {
	switch method {
	case http.MethodGet:
		return pr.mux.Get
	case http.MethodPost:
		return pr.mux.Post
	case http.MethodPut:
		return pr.mux.Put
	case http.MethodDelete:
		return pr.mux.Delete
	case http.MethodHead:
		return pr.mux.Head
	case http.MethodOptions:
		return pr.mux.Options
	case http.MethodPatch:
		return pr.mux.Patch
	default:
		panic(ErrInvalidMethod)
	}
}

func (pr *chiRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.mux.ServeHTTP(w, r)
}

func (pr *chiRouter) SetNotFoundHandler(handler http.Handler) {
	pr.mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

func (pr *chiRouter) SetNotAllowedHandler(handler http.Handler) {
	pr.mux.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

func validMethod(method string) bool {
	return method == http.MethodDelete || method == http.MethodGet ||
		method == http.MethodHead || method == http.MethodOptions ||
		method == http.MethodPatch || method == http.MethodPost ||
		method == http.MethodPut
}
