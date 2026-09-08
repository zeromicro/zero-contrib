package nacos

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/resolver"
)

type resolvr struct {
	cancelFunc context.CancelFunc
	cli        interface {
		Unsubscribe(*vo.SubscribeParam) error
		CloseClient()
	}
	subParam  vo.SubscribeParam
	pipe      chan []string
	closeOnce sync.Once
}

func (r *resolvr) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver, unsubscribes from Nacos, shuts down the naming
// client, and drains the internal pipe channel so the watcher goroutine can't
// block on a send after the consumer has stopped.
func (r *resolvr) Close() {
	r.closeOnce.Do(func() {
		r.cancelFunc()
		_ = r.cli.Unsubscribe(&r.subParam)
		r.cli.CloseClient()
		// drain so any in-flight CallBackHandle send doesn't block forever
		for {
			select {
			case <-r.pipe:
			default:
				return
			}
		}
	})
}

type watcher struct {
	ctx context.Context
	out chan<- []string
}

func newWatcher(ctx context.Context, out chan<- []string) *watcher {
	return &watcher{ctx: ctx, out: out}
}

func (nw *watcher) CallBackHandle(services []model.Instance, err error) {
	if err != nil {
		logger.Error("[Nacos resolver] watcher call back handle error:%v", err)
		return
	}
	ee := make([]string, 0, len(services))
	for _, s := range services {
		if s.Metadata != nil && s.Metadata["gRPC_port"] != "" {
			ee = append(ee, fmt.Sprintf("%s:%s", s.Ip, s.Metadata["gRPC_port"]))
		} else {
			ee = append(ee, fmt.Sprintf("%s:%d", s.Ip, s.Port))
		}
	}
	select {
	case nw.out <- ee:
	case <-nw.ctx.Done():
	}
}

func populateEndpoints(ctx context.Context, clientConn resolver.ClientConn, input <-chan []string) {
	for {
		select {
		case cc := <-input:
			connsSet := make(map[string]struct{}, len(cc))
			for _, c := range cc {
				connsSet[c] = struct{}{}
			}
			conns := make([]resolver.Address, 0, len(connsSet))
			for c := range connsSet {
				conns = append(conns, resolver.Address{Addr: c})
			}
			sort.Sort(byAddressString(conns)) // Don't replace the same address list in the balancer
			_ = clientConn.UpdateState(resolver.State{Addresses: conns})
		case <-ctx.Done():
			logx.Info("[Nacos resolver] Watch has been finished")
			return
		}
	}
}

// byAddressString sorts resolver.Address by Address Field  sorting in increasing order.
type byAddressString []resolver.Address

func (p byAddressString) Len() int           { return len(p) }
func (p byAddressString) Less(i, j int) bool { return p[i].Addr < p[j].Addr }
func (p byAddressString) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
