package consul

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/core/proc"
)

func RegisterService(listenOn string, c Conf) error {
	pubListenOn := figureOutListenOn(listenOn)

	host, ports, err := net.SplitHostPort(pubListenOn)
	if err != nil {
		return fmt.Errorf("failed parsing address error: %v", err)
	}
	port, _ := strconv.ParseUint(ports, 10, 16)

	client, err := api.NewClient(&api.Config{Scheme: "http", Address: c.Host})
	if err != nil {
		return fmt.Errorf("create consul client error: %v", err)
	}
	// 服务节点的名称
	serviceID := fmt.Sprintf("%s-%s-%d", c.Key, host, port)

	reg := &api.AgentServiceRegistration{
		ID:      serviceID, // 服务节点的名称
		Name:    c.Key,     // 服务名称
		Tags:    c.Tag,     // tag，可以为空
		Meta:    c.Meta,    // meta， 可以为空
		Port:    int(port), // 服务端口
		Address: host,      // 服务 IP
		Checks: []*api.AgentServiceCheck{ // 健康检查
			{
				CheckID:                        serviceID, // 服务节点的名称
				TTL:                            "30s",     // 健康检查间隔
				Status:                         "passing",
				DeregisterCriticalServiceAfter: "90s", // 注销时间，相当于过期时间
			},
		},
	}

	if err := client.Agent().ServiceRegister(reg); err != nil {
		return fmt.Errorf("initial register service '%s' host to consul error: %s", c.Key, err.Error())
	}

	// initial register service check
	check := api.AgentServiceCheck{TTL: "20s", Status: "passing", DeregisterCriticalServiceAfter: "90s"}
	err = client.Agent().CheckRegister(&api.AgentCheckRegistration{ID: serviceID, Name: c.Key, ServiceID: serviceID, AgentServiceCheck: check})
	if err != nil {
		return fmt.Errorf("initial register service check to consul error: %s", err.Error())
	}

	// routine to update ttl
	go func() {
		ticker := time.NewTicker(time.Second * 20)
		defer ticker.Stop()
		for {
			<-ticker.C
			err = client.Agent().UpdateTTL(serviceID, "", "passing")
			logx.Info("update ttl")
			if err != nil {
				logx.Infof("update ttl of service error: %v", err.Error())
			}
		}
	}()
	// consul deregister
	proc.AddShutdownListener(func() {
		err := client.Agent().ServiceDeregister(serviceID)
		if err != nil {
			logx.Info("deregister service error: ", err.Error())
		} else {
			logx.Info("deregistered service from consul server.")
		}
		err = client.Agent().CheckDeregister(serviceID)
		if err != nil {
			logx.Info("deregister check error: ", err.Error())
		}
	})

	return nil
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
