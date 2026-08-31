### Quick Start

Prerequesites:

- Install `go-zero`:

```console
go get -u github.com/zeromicro/go-zero
```

Download the module:

```console
go get -u github.com/wqyjh/zero-contrib/router/gin
```

For example:

```go
package main

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	zerogin "github.com/wqyjh/zero-contrib/router/gin"
	"github.com/gin-gonic/gin"
)

type CommonPathID struct {
	ID   int    `path:"id"`
	Name string `path:"name"`
}

func (c *CommonPathID) String() string {
	var builder strings.Builder
	builder.WriteString("CommonPathID(")
	builder.WriteString(fmt.Sprintf("ID=%v", c.ID))
	builder.WriteString(fmt.Sprintf(", Name=%s", c.Name))
	builder.WriteByte(')')
	return builder.String()
}

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	server := gin.New()
	r := zerogin.NewRouter(r)
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
		Path:   "/api/:name/:id",  // GET /api/joh/123
		Handler: func(w http.ResponseWriter, r *http.Request) {
			var commonPath CommonPathID
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
