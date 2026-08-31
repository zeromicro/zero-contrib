package gin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	Middleware func(next http.Handler) http.Handler
)

func HandlerFunc(ctx *gin.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx.Request = r
		ctx.Next()
	}
}

func ZeroMiddleware(middleware Middleware) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		next := HandlerFunc(ctx)
		middleware(next).ServeHTTP(ctx.Writer, ctx.Request)
	}
}
