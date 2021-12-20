package main

import (
	"context"
	"flag"
	"etcd-example/3-etcd-grpc-lb/etcdv3"
	"etcd-example/3-etcd-grpc-lb/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

var port = flag.String("port", ":8001", "server port")

type helloService struct{}

func (h helloService) Say(ctx context.Context, request *proto.HelloRequest) (*proto.HelloReply, error) {
	log.Println("receive:" + request.Name)
	return &proto.HelloReply{
		Code:  200,
		Value: "hello " + request.Name,
	}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", *port)
	if err != nil {
		panic(err)
	}

	// 新建gRPC服务实例
	grpcServer := grpc.NewServer()

	// 在gRPC服务中注册我们的服务
	proto.RegisterHelloServiceServer(grpcServer, &helloService{})

	// 把服务注册到etcd
	service, err := etcdv3.NewServiceRegister([]string{"localhost:2379"}, "grpclb_test1", "localhost"+*port, 5)
	if err != nil {
		panic(err)
	}
	defer service.Close()

	log.Println("server starting on " + *port)
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
