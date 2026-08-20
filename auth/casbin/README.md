### Quick Start

Prerequisites:

* Install `go-zero`: go get -u github.com/zeromicro/go-zero

Download the module:

```shell
go get -u github.com/wqyjh/zero-contrib/auth/casbin
```

For example:

```go
package main

import (
	stdcasbin "github.com/casbin/casbin/v2"
	"github.com/zeromicro/go-zero/rest"
	"github.com/wqyjh/zero-contrib/auth/casbin"
)

func main() {
	// load the casbin model and policy from files, database is also supported.
	e, _ := stdcasbin.NewEnforcer("auth_model.conf", "auth_policy.csv")

	// define your router, and use the Casbin auth middleware.
	// the access that is denied by auth will return HTTP 403 error.
	// set username as the user unique identity field.
	authorizer := casbin.NewAuthorizer(e, casbin.WithUidField("username"))
	conf := rest.RestConf{}
	server := rest.MustNewServer(conf)
	server.Use(rest.ToMiddleware(authorizer))
}

```

documentation:

The authorization determines a request based on {subject, object, action}, which means what subject can perform what
action on what object. In this plugin, the meanings are:

- subject: It comes from user unique identity in JWT Claims.
- object: the URL path for the web resource like "dataset1/item1".
- action: HTTP method like GET, POST, PUT, DELETE, or the high-level actions you defined like "read-file", "write-blog".

For how to write authorization policy and other details, please refer
to [the Casbin's documentation](https://github.com/casbin/casbin).

gratitude:

- https://github.com/gin-contrib/authz