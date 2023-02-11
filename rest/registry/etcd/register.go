package etcd

import (
	"fmt"
	"os"
	"strings"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/rest"
)

const (
	allEths  = "0.0.0.0"
	envPodIP = "POD_IP"
)

// RegisterRest register reset to etcd.
func RegisterRest(etcd discov.EtcdConf, svrConf rest.RestConf) error {
	if err := etcd.Validate(); err != nil {
		return err
	}

	listenOn := fmt.Sprintf("%s:%d", svrConf.Host, svrConf.Port)
	pubListenOn := figureOutListenOn(listenOn)
	var pubOpts []discov.PubOption
	if etcd.HasAccount() {
		pubOpts = append(pubOpts, discov.WithPubEtcdAccount(etcd.User, etcd.Pass))
	}
	if etcd.HasTLS() {
		pubOpts = append(pubOpts, discov.WithPubEtcdTLS(etcd.CertFile, etcd.CertKeyFile,
			etcd.CACertFile, etcd.InsecureSkipVerify))
	}

	pubClient := discov.NewPublisher(etcd.Hosts, etcd.Key, pubListenOn, pubOpts...)
	proc.AddShutdownListener(func() {
		pubClient.Stop()
	})

	return pubClient.KeepAlive()
}

func figureOutListenOn(listenOn string) string {
	fields := strings.Split(listenOn, ":")
	if len(fields) == 0 {
		return listenOn
	}

	host := fields[0]
	if len(host) > 0 && host != allEths {
		return listenOn
	}

	ip := os.Getenv(envPodIP)
	if len(ip) == 0 {
		ip = netx.InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}

	return strings.Join(append([]string{ip}, fields[1:]...), ":")
}
