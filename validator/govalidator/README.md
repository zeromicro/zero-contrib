### Quick Start

## Prerequesites:

- Install `go-zero`:

```shell
go get -u github.com/zeromicro/go-zero
```



+ Download the module:

```shell
go get -u github.com/zeromicro/zero-contrib/validator/govalidator
```

## For example:

```go
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.CommonReply, error) {
	// todo: add your logic here and delete this line
  ....
	err := govalidator.DefaultGetValidParams(l.ctx, in)
	if err != nil {
		return nil, errorx.NewDefaultError(fmt.Sprintf("validate校验错误: %v", err))
	}
	.....

}

```

