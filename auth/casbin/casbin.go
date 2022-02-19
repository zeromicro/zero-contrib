package casbin

import (
	"github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
)

// Authorizer stores the casbin handler
type Authorizer struct {
	enforcer *casbin.Enforcer
}

// NewAuthorizer returns the authorizer, uses a Casbin enforcer as input
func NewAuthorizer(e *casbin.Enforcer) func(http.Handler) http.Handler {
	a := &Authorizer{enforcer: e}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !a.CheckPermission(request) {
				a.RequirePermission(writer)
			}
			next.ServeHTTP(writer, request)
		})
	}
}

// GetUserName gets the username from the request.
// Currently, only HTTP basic authentication is supported
func (a *Authorizer) GetUserName(r *http.Request) (string, bool) {
	username, ok := r.Context().Value("username").(string)

	return username, ok
}

// CheckPermission checks the user/method/path combination from the request.
// Returns true (permission granted) or false (permission forbidden)
func (a *Authorizer) CheckPermission(r *http.Request) bool {
	user, ok := a.GetUserName(r)
	if !ok {
		return false
	}
	method := r.Method
	path := r.URL.Path

	allowed, err := a.enforcer.Enforce(user, path, method)
	if err != nil {
		logx.WithContext(r.Context()).Errorf("[CASBIN] enforce err %s", err.Error())
	}

	return allowed
}

// RequirePermission returns the 403 Forbidden to the client.
func (a *Authorizer) RequirePermission(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusForbidden)

}
