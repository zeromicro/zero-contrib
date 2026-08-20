### Quick Start
Prerequesites:
* Install `go-zero`: go get -u github.com/zeromicro/go-zero@master


Download the module:
```console
go get -u github.com/wqyjh/zero-contrib/router/chi
```

For example:

```go
package main

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
    "github.com/wqyjh/zero-contrib/router/chi"
	"net/http"
	"strings"
)

type CommonPath struct {
	Year  int `path:"year"`
	Month int `path:"month"`
	Day   int `path:"day"`
}

func (c *CommonPath) String() string {
	var builder strings.Builder
	builder.WriteString("CommonPath(")
	builder.WriteString(fmt.Sprintf("Year=%v", c.Year))
	builder.WriteString(fmt.Sprintf(", Month=%v", c.Month))
	builder.WriteString(fmt.Sprintf(", Day=%v", c.Day))
	builder.WriteByte(')')
	return builder.String()
}

func main() {
    r := chi.NewRouter()
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
		w.Write([]byte("nothing here"))
	}))
	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
    r.SetNotAllowedHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(405)
		w.Write([]byte("405"))
	}))
	defer engine.Stop()

	engine.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/api/{month}-{day}-{year}",  // GET /articles/01-16-2017
		Handler: func(w http.ResponseWriter, r *http.Request) {
			var commonPath CommonPath
			err := httpx.Parse(r, &commonPath)
			if err != nil {
				return
			}
			w.Write([]byte(commonPath.String()))

		},
	})
	engine.Start()
}

```
