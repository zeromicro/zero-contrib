### Quick Start

Prerequisites:

Download the module:

```console
go get -u github.com/wqyjh/zero-contrib/zrpc/registry/consul
```

For example:

## 修改RPC服务的代码

- etc/\*.yaml

```yaml
Consul:
  Host: 127.0.0.1:8500 # consul endpoint
  Token: 'f0512db6-76d6-f25e-f344-a98cc3484d42' # consul ACL token (optional)
  Key: add.rpc # service name registered to Consul
  Meta:
    Protocol: grpc
  Tag:
    - tag
    - rpc
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
import _ "github.com/wqyjh/zero-contrib/zrpc/registry/consul"

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
	
	})
 	// register service to consul
	_ = consul.RegisterService(c.ListenOn, c.Consul)

	server.Start()
}
```

### ACL Token

the token need contain the **service** policy of `"write"`

```
service "add.rpc" {
  policy = "write"
}
service "check.rpc" {
  policy = "write"
}
```

## 修改API服务里的代码

- main.go

```go
import _ "github.com/wqyjh/zero-contrib/zrpc/registry/consul"
```

- etc/\*.yaml

```yaml
# consul://[user:passwd]@host/service?param=value'
# format like this
Add:
  Target: consul://127.0.0.1:8500/add.rpc?wait=14s
Check:
  Target: consul://127.0.0.1:8500/check.rpc?wait=14s
  
# ACL token support
Add:
  Target: consul://127.0.0.1:8500/add.rpc?wait=14s&token=f0512db6-76d6-f25e-f344-a98cc3484d42
Check:
  Target: consul://127.0.0.1:8500/check.rpc?wait=14s&token=f0512db6-76d6-f25e-f344-a98cc3484d42
```

### ACL Token

the token need contain the **node** and **service** policy of `"read"`

```
node "consul-server" {
  policy = "read"
}
service "add.rpc" {
	policy = "read"
}
service "check.rpc" {
	policy = "read"
}
```