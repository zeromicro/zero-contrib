package gin

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

type ginRouter struct {
	g *gin.Engine
}

// NewRouter returns a gin.Router.
func NewRouter(g *gin.Engine, opts ...Option) httpx.Router {
	cfg := config{
		redirectTrailingSlash: true,
		redirectFixedPath:     false,
	}
	cfg.options(opts...)

	g.RedirectTrailingSlash = cfg.redirectTrailingSlash
	g.RedirectFixedPath = cfg.redirectFixedPath
	return &ginRouter{g: g}
}

func (pr *ginRouter) Handle(method, reqPath string, handler http.Handler) error {
	if !validMethod(method) {
		return ErrInvalidMethod
	}

	if len(reqPath) == 0 || reqPath[0] != '/' {
		return ErrInvalidPath
	}

	pr.g.Handle(strings.ToUpper(method), reqPath, func(ctx *gin.Context) {
		params := make(map[string]string)
		for i := 0; i < len(ctx.Params); i++ {
			params[ctx.Params[i].Key] = ctx.Params[i].Value
		}
		if len(params) > 0 {
			ctx.Request = pathvar.WithVars(ctx.Request, params)
		}
		handler.ServeHTTP(ctx.Writer, ctx.Request)
	})
	return nil
}

func (pr *ginRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pr.g.ServeHTTP(w, r)
}

func (pr *ginRouter) SetNotFoundHandler(handler http.Handler) {
	pr.g.NoRoute(gin.WrapH(handler))
}

func (pr *ginRouter) SetNotAllowedHandler(handler http.Handler) {
	pr.g.HandleMethodNotAllowed = true
	pr.g.NoMethod(gin.WrapH(handler))
}

func validMethod(method string) bool {
	return method == http.MethodDelete || method == http.MethodGet ||
		method == http.MethodHead || method == http.MethodOptions ||
		method == http.MethodPatch || method == http.MethodPost ||
		method == http.MethodPut
}
