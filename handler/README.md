## Zero Handler Contrib List

ETag: ETag Handler to support [ETag](https://en.wikipedia.org/wiki/HTTP_ETag) both for weak and strong validation

## ETag

### Prerequisites:

* Install `go-zero`: go get -u github.com/zeromicro/go-zero

### Download the module:

```shell
go get -u github.com/wqyjh/zero-contrib/handler
```

### Example:

```go
package api

import (
	...
	
	"github.com/zeromicro/go-zero/rest"
	"github.com/wqyjh/zero-contrib/handler"
)

func main() {
	...

	server := rest.MustNewServer(c.RestConf)
	server.Use(handler.NewETagMiddleware(true).Handle)
}

```

### Gratitude:

- https://github.com/go-http-utils/etag