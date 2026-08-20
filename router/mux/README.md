### Quick Start

Prerequesites:

- Install `go-zero`:

```console
go get -u github.com/zeromicro/go-zero
```

Download the module:

```console
go get -u github.com/wqyjh/zero-contrib/router/mux
```

For example:

```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/wqyjh/zero-contrib/router/mux"
	"net/http"
)

type User struct {
	Id   int64  `path:"id"`   //`form:"id"`   |  `json:"id"`
	Name string `path:"name"` //`form:"name"` |  `json:"name"`
}

func main() {
	r := mux.NewRouter()
	engine := rest.MustNewServer(rest.RestConf{
		ServiceConf: service.ServiceConf{
			Log: logx.LogConf{
				Mode: "console",
			},
		},
		Port:     3345,
		Timeout:  20000,
		MaxConns: 500,
	}, rest.WithRouter(r))

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	r.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		w.Write([]byte("404 callback here"))
	}))
	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
	r.SetNotAllowedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("405 callback here"))
	}))
	defer engine.Stop()

	engine.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/{name}/{id}", // GET /api/joh/123
		Handler: func(w http.ResponseWriter, r *http.Request) {
			var user User
			err := httpx.Parse(r, &user)
			if err != nil {
				return
			}
			userStr, _ := json.Marshal(user)
			fmt.Println(userStr)
			w.Write([]byte(userStr))
		},
	})
	engine.Start()
}
```
