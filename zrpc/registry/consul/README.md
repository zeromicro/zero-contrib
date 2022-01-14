### Quick Start

Prerequesites:

Download the module:

```console
go get -u github.com/zeromicro/zero-contrib/zrpc/registry/consul
```

For example:

## Service

- etc/\*.yaml

```yaml
  Consul:
  Host: 192.168.100.15:8500
  Key:consul.rpc
  Meta:
    Protocol: grpc
  Tag:
    -
      tag
      rpc

```

- internal/config/config.go

```go
type Config struct {
	zrpc.RpcServerConf
	Consul consul.Conf
}
```

- main.go

```go
import _ "github.com/zeromicro/zero-contrib/zrpc/registry/consul"

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {

	})
	// 注册服务
	_ = consul.RegitserService(c.ListenOn, c.Consul)

	server.Start()
}
```

## Client

- main.go

```go
import _ "github.com/zeromicro/zero-contrib/zrpc/registry/consul"
```

- etc/\*.yaml

```yaml
# consul://[user:passwd]@host/service?param=value'
Target: consul://192.168.100.15:8500/consul.rpc?wait=14s
```
