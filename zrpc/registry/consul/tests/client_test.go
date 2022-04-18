package tests

import (
	"testing"
	"time"

	_ "github.com/zeromicro/zero-contrib/zrpc/registry/consul"
	"google.golang.org/grpc"
)

func TestCLient(t *testing.T) {
	conn, err := grpc.Dial("consul://127.0.0.1:8500/gozero?wait=14s&tag=public", grpc.WithInsecure(), grpc.WithBalancerName("round_robin"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	time.Sleep(29 * time.Second)
}
