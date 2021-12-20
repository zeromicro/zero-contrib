### Quick Start

Prerequesites:

Download the module:

```console
go get -u github.com/zeromicro/zero-contrib/zrpc/registry/polaris
```

For example:

## Service

- ./polaris.yaml

```yaml
global:
  serverConnector:
    addresses:
      - 127.0.0.1:8091
```

- main.go

```go
import _ "github.com/zeromicro/zero-contrib/zrpc/registry/polaris"

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {

	})
	// 注册服务
	_ = polaris.RegitserService(c.ListenOn, 
                polaris.WithServiceName("EchoServerZero"),
                polaris.WithNamespace("default"),
            )

	server.Start()
}
```

## Client

- main.go

```go
import _ "github.com/zeromicro/zero-contrib/zrpc/registry/polaris"
```

- etc/\*.yaml

```yaml
# polaris://[user:passwd]@host/service?param=value'
Target: polaris://127.0.0.1:8091/EchoServerZero?wait=14s
```
