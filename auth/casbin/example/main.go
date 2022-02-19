package main

import (
	stdcasbin "github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/zero-contrib/auth/casbin"
)

func main() {
	// load the casbin model and policy from files, database is also supported.
	e, _ := stdcasbin.NewEnforcer("auth_model.conf", "auth_policy.csv")

	// define your router, and use the Casbin auth middleware.
	// the access that is denied by auth will return HTTP 403 error.
	authorizer := casbin.NewAuthorizer(e)
	conf := rest.RestConf{}
	server := rest.MustNewServer(conf)
	server.Use(rest.ToMiddleware(authorizer))
}
