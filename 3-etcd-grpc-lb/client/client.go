package main

import (
	"context"
	"etcd-example/3-etcd-grpc-lb/etcdv3"
	"etcd-example/3-etcd-grpc-lb/proto"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/resolver"
	"log"
	"strconv"
)

func main() {
	builder := etcdv3.NewServiceDiscovery([]string{"localhost:2379"})
	resolver.Register(builder)

	conn, err := grpc.Dial(
		builder.Scheme()+":///"+"grpclb_test1",
		// grp负载均衡策略，默认为pick_first，这里选择round_robin，当然可以自定义负载均衡策略
		//grpc.WithBalancerName("这里选择round_robin"),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
		grpc.WithInsecure(),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	grpcClient := proto.NewHelloServiceClient(conn)
	ctx := context.Background()
	for i := 1; i <= 20; i++ {
		resp, err := grpcClient.Say(ctx, &proto.HelloRequest{
			Name: "grpc" + strconv.Itoa(i),
		})
		if err != nil {
			panic(err)
		}
		log.Println(resp)
	}
}
