package consul

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/jpillora/backoff"
	"github.com/tal-tech/go-zero/core/logx"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&builder{})
}

type resolvr struct {
	cancelFunc context.CancelFunc
}

func (r *resolvr) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (r *resolvr) Close() {
	r.cancelFunc()
}

type servicer interface {
	Service(string, string, bool, *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error)
}

func watchConsulService(ctx context.Context, s servicer, tgt target, out chan<- []string) {
	res := make(chan []string)
	quit := make(chan struct{})
	bck := &backoff.Backoff{
		Factor: 2,
		Jitter: true,
		Min:    10 * time.Millisecond,
		Max:    tgt.MaxBackoff,
	}
	go func() {
		var lastIndex uint64
		for {
			ss, meta, err := s.Service(
				tgt.Service,
				tgt.Tag,
				tgt.Healthy,
				&api.QueryOptions{
					WaitIndex:         lastIndex,
					Near:              tgt.Near,
					WaitTime:          tgt.Wait,
					Datacenter:        tgt.Dc,
					AllowStale:        tgt.AllowStale,
					RequireConsistent: tgt.RequireConsistent,
				},
			)
			if err != nil {
				logx.Errorf("[Consul resolver] Couldn't fetch endpoints. target={%s}; error={%v}", tgt.String(), err)

				time.Sleep(bck.Duration())
				continue
			}

			bck.Reset()
			lastIndex = meta.LastIndex
			logx.Infof("[Consul resolver] %d endpoints fetched in(+wait) %s for target={%s}",
				len(ss),
				meta.RequestTime,
				tgt.String(),
			)

			ee := make([]string, 0, len(ss))
			for _, s := range ss {
				address := s.Service.Address
				if s.Service.Address == "" {
					address = s.Node.Address
				}
				ee = append(ee, fmt.Sprintf("%s:%d%s%s", address, s.Service.Port, dyeingStp, strings.Join(s.Service.Tags, ",")))
			}

			if tgt.Limit != 0 && len(ee) > tgt.Limit {
				ee = ee[:tgt.Limit]
			}
			select {
			case res <- ee:
				continue
			case <-quit:
				return
			}
		}
	}()

	for {
		select {
		case ee := <-res:
			out <- ee
		case <-ctx.Done():
			close(quit)
			return
		}
	}
}

func populateEndpoints(ctx context.Context, clientConn resolver.ClientConn, input <-chan []string) {
	for {
		select {
		case cc := <-input:
			connsSet := make(map[string]string, len(cc))
			for _, c := range cc {
				addr, tag := splitAddr(c)
				connsSet[addr] = tag
			}
			conns := make([]resolver.Address, 0, len(connsSet))
			for c, tags := range connsSet {
				rAddr := resolver.Address{Addr: c}
				if tags != "" {
					rAddr.Attributes = attributes.New(dyeingKey, tags)
				}
				conns = append(conns, rAddr)
			}
			sort.Sort(byAddressString(conns)) // Don't replace the same address list in the balancer
			_ = clientConn.UpdateState(resolver.State{Addresses: conns})
		case <-ctx.Done():
			logx.Info("[Consul resolver] Watch has been finished")
			return
		}
	}
}

func splitAddr(conn string) (addr, tag string) {
	connAttrs := strings.Split(conn, dyeingStp)
	if len(connAttrs) < 2 {
		return connAttrs[0], ""
	}

	if connAttrs[1] == "" {
		return connAttrs[0], ""
	}

	for _, tag := range strings.Split(connAttrs[1], ",") {
		if strings.HasPrefix(tag, dyeingPrefix) {
			return connAttrs[0], tag[len(dyeingPrefix):]
		}
	}
	return connAttrs[0], ""
}

// byAddressString sorts resolver.Address by Address Field  sorting in increasing order.
type byAddressString []resolver.Address

func (p byAddressString) Len() int           { return len(p) }
func (p byAddressString) Less(i, j int) bool { return p[i].Addr < p[j].Addr }
func (p byAddressString) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
