### Quick Start

Prerequisites:

Download the module:

```console
go get -u github.com/wqyjh/zero-contrib/zrpc/registry/polaris
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
import _ "github.com/wqyjh/zero-contrib/zrpc/registry/polaris"

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {

	})
	// register service to polaris
    opts := polaris.NewPolarisConfig(c.ListenOn)
	opts.ServiceName = "EchoServerZero" 
	opts.Namespace = "default"
	opts.ServiceToken = "2af8fdf2534f451e8f01881d1b66f9ec"
    _ = polaris.RegisterService(opts)

	server.Start()
}
```

## Client

- main.go

```go
import _ "github.com/wqyjh/zero-contrib/zrpc/registry/polaris"
```

- etc/\*.yaml

```yaml
# polaris://[user:passwd]@host/service?param=value'
Target: polaris://127.0.0.1:8091/EchoServerZero?wait=14s
```
