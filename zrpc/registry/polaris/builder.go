package polaris

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"google.golang.org/grpc/resolver"

	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/config"
	"github.com/polarismesh/polaris-go/pkg/model"
)

var (
	consumers map[string]api.ConsumerAPI = make(map[string]api.ConsumerAPI)
	lock      *sync.Mutex                = &sync.Mutex{}
)

func init() {
	resolver.Register(&builder{})
}

// builder implements resolver.Builder and use for constructing all consul resolvers
type builder struct{}

func (b *builder) Build(url resolver.Target, conn resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	dsn := strings.Join([]string{schemeName + ":/", url.Authority, url.Endpoint}, "/")
	tgr, err := parseURL(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "Wrong polaris URL")
	}

	var polarisErr error

	func() {
		lock.Lock()
		defer lock.Unlock()

		if _, exist := consumers[tgr.Addr]; exist {
			return
		}

		sdkCtx, err := api.InitContextByConfig(config.NewDefaultConfiguration([]string{tgr.Addr}))
		if err != nil {
			polarisErr = errors.Wrap(err, "Fail init polaris SDKContext")
			return
		}
		consumerAPI := api.NewConsumerAPIByContext(sdkCtx)
		consumers[tgr.Addr] = consumerAPI
	}()

	if polarisErr != nil {
		return nil, polarisErr
	}

	consumerAPI := consumers[tgr.Addr]
	ctx, cancel := context.WithCancel(context.Background())
	pipe := make(chan []string, 4)

	go newWatcher(pipe).startWatch(ctx, consumerAPI, &api.WatchServiceRequest{
		WatchServiceRequest: model.WatchServiceRequest{
			Key: model.ServiceKey{
				Namespace: tgr.Namespace,
				Service:   tgr.Service,
			},
		},
	})

	go populateEndpoints(ctx, conn, pipe)

	return &resolvr{cancelFunc: cancel}, nil
}

// Scheme returns the scheme supported by this resolver.
// Scheme is defined at https://github.com/grpc/grpc/blob/master/doc/naming.md.
func (b *builder) Scheme() string {
	return schemeName
}
