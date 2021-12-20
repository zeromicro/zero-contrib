package polaris

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/polarismesh/polaris-go/api"
	"github.com/polarismesh/polaris-go/pkg/model"
	"github.com/tal-tech/go-zero/core/logx"
	"google.golang.org/grpc/resolver"
)

type resolvr struct {
	cancelFunc context.CancelFunc
}

func (r *resolvr) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (r *resolvr) Close() {
	r.cancelFunc()
}

type polarisServiceWatcher struct {
	out chan<- []string
}

func newWatcher(out chan<- []string) *polarisServiceWatcher {
	return &polarisServiceWatcher{
		out: out,
	}
}

func (watcher *polarisServiceWatcher) startWatch(ctx context.Context, consumer api.ConsumerAPI, subscribeParam *api.WatchServiceRequest) {
	for {
		resp, err := consumer.WatchService(subscribeParam)
		if err != nil {
			time.Sleep(time.Duration(500 * time.Millisecond))
			continue
		}

		instances := resp.GetAllInstancesResp.Instances
		ee := make([]string, len(instances)+1)
		for i := range instances {
			ins := instances[i]
			ee[i] = fmt.Sprintf("%s:%d", ins.GetHost(), ins.GetPort())
		}
		if len(ee) != 0 {
			watcher.out <- ee
		}

		logx.Info("[Polaris resolver] Watch has been start, param : %#v", subscribeParam)

		select {
		case <-ctx.Done():
			logx.Info("[Polaris resolver] Watch has been finished")
			return
		case event := <-resp.EventChannel:
			eType := event.GetSubScribeEventType()
			if eType == api.EventInstance {
				insEvent := event.(*model.InstanceEvent)
				ee := make([]string, len(insEvent.AddEvent.Instances)+len(insEvent.UpdateEvent.UpdateList))
				if insEvent.AddEvent != nil {
					for _, s := range insEvent.AddEvent.Instances {
						ee = append(ee, fmt.Sprintf("%s:%d", s.GetHost(), s.GetPort()))
					}
				}
				if insEvent.UpdateEvent != nil {
					for _, s := range insEvent.UpdateEvent.UpdateList {
						ee = append(ee, fmt.Sprintf("%s:%d", s.After.GetHost(), s.After.GetPort()))
					}
				}

				if len(ee) != 0 {
					watcher.out <- ee
				}
			}
		}
	}
}

// populateEndpoints
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
			logx.Info("[Polaris resolver] Watch has been finished")
			return
		}
	}
}

// byAddressString sorts resolver.Address by Address Field  sorting in increasing order.
type byAddressString []resolver.Address

func (p byAddressString) Len() int           { return len(p) }
func (p byAddressString) Less(i, j int) bool { return p[i].Addr < p[j].Addr }
func (p byAddressString) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
