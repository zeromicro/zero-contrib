package mux

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

var (
	// ErrInvalidMethod is an error that indicates not a valid http method.
	ErrInvalidMethod = errors.New("not a valid http method")
	// ErrInvalidPath is an error that indicates path is not start with /.
	ErrInvalidPath = errors.New("path must begin with '/'")
)

type muxRouter struct {
	g *mux.Router
}

// NewRouter returns a mux.Router.
func NewRouter() httpx.Router {
	g := mux.NewRouter()
	return &muxRouter{g: g}
}

func (pr *muxRouter) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) {
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' {
		return ErrInvalidPath
	}

	pr.g.HandleFunc(reqPath, func(w http.ResponseWriter, r *http.Request) {
		params := make(map[string]string)
		vars := mux.Vars(r)
		for key, val := range vars {
			params[key] = val
		}
		if len(params) > 0 {
			r = pathvar.WithVars(r, params)
		}
		handler.ServeHTTP(w, r)
	}).Methods(strings.ToUpper(method))
	return nil
}

func (pr *muxRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.g.ServeHTTP(w, r)
}

func (pr *muxRouter) SetNotFoundHandler(handler http.Handler) {
	pr.g.NotFoundHandler = handler
}

func (pr *muxRouter) SetNotAllowedHandler(handler http.Handler) {
	pr.g.MethodNotAllowedHandler = handler
}

func validMethod(method string) bool {
	return method == http.MethodDelete || method == http.MethodGet ||
		method == http.MethodHead || method == http.MethodOptions ||
		method == http.MethodPatch || method == http.MethodPost ||
		method == http.MethodPut
}
