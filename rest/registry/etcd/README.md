### Quick Start

Prerequisites:

Download the module:

```console
go get -u github.com/wqyjh/zero-contrib/rest/registry/etcd
```

For example:

## 修改REST服务的代码

- etc/*.yaml

```yaml
Name: user.api
Host: 0.0.0.0
Port: 8888
Etcd:
  Hosts:
    - etcd:2379
  Key: user.api

```

- internal/config/config.go

```go
type Config struct {
  rest.RestConf
  Etcd discov.EtcdConf // etcd register center config
}
```

- main.go

```go
package main

import (
	"flag"

	"github.com/zeromicro/go-zero/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/wqyjh/zero-contrib/rest/registry/etcd"
)

var configFile = flag.String("f", "etc/user-api.yaml", "the config file")

func main() {
  flag.Parse()

  var c Config
  conf.MustLoad(*configFile, &c)

  server := rest.MustNewServer(c.RestConf)
  // register rest to etcd
  logx.Must(etcd.RegisterRest(c.Etcd, c.RestConf))

  server.Start()
}
```