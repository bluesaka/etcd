package main

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/client/v3"
)

// ServiceRegister 服务注册
type ServiceRegister struct {
	cli           *clientv3.Client
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
}

// NewServiceRegister 注册服务
func NewServiceRegister(endpoints []string, key, val string, lease int64) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	ser := &ServiceRegister{
		cli: cli,
		key: key,
		val: val,
	}

	if err := ser.putKeyWithLease(lease); err != nil {
		return nil, err
	}

	return ser, nil
}

// 设置key并绑定租约
func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 新建租约
	resp, err := s.cli.Grant(ctx, lease)
	if err != nil {
		return err
	}

	// 设置key并绑定租约
	_, err = s.cli.Put(ctx, s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// 设置KeepAlive自动续租，注意不能使用上面的ctx，KeepAlive续租的context不能关闭
	keepAliveChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}

	s.keepAliveChan = keepAliveChan
	log.Printf("put key:%s val:%s success\n", s.key, s.val)
	return nil
}

// ListenKeepAlive 监听自动续租
func (s *ServiceRegister) ListenKeepAlive() {
	//for {
	//	ka := <-s.keepAliveChan
	//	log.Printf("续租成功 keepAliveChan: %v, ttl: %d", ka, ka.TTL)
	//}
	for ka := range s.keepAliveChan {
		log.Printf("续租成功: %v, ttl: %d", ka, ka.TTL)
	}
	log.Println("关闭续租")
}

func main() {
	s := Register("/HTTP/s1", "localhost:8081", 5)
	Register("/gRPC/s2.rpc", "localhost:8082", 6)
	Register("/HTTP2/s1", "localhost:28082", 6)

	// 监听自动续租
	go s.ListenKeepAlive()
	select {}
}

// Register 注册服务
func Register(key, val string, lease int64) *ServiceRegister {
	s, err := NewServiceRegister([]string{"localhost:2379"}, key, val, lease)
	if err != nil {
		log.Fatal(err)
	}
	return s
}
