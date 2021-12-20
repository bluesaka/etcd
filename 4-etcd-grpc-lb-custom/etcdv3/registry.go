package etcdv3

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"log"
	"strconv"
	"time"
)

// ServiceRegister 租约注册服务
type ServiceRegister struct {
	client        *clientv3.Client //etcd client
	leaseID       clientv3.LeaseID //租约ID
	key           string
	value         string
	weight        int
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
}

func NewServiceRegister(endpoints []string, serverName, addr string, weight int, lease int64) (service *ServiceRegister, err error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 2 * time.Second,
	})
	if err != nil {
		return
	}

	// 使用Status来检测endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err = client.Status(ctx, endpoints[0])
	if err != nil {
		log.Printf("endpoint: %s is not available\n", endpoints[0])
		panic(err)
	}

	service = &ServiceRegister{
		client: client,
		key:    "/" + schema + "/" + serverName + "/" + addr,
		weight: weight,
	}

	err = service.putKeyWithLease(lease)
	if err != nil {
		return
	}
	return
}

// 设置key并绑定租约
func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	// 新建租约
	ctx := context.Background()
	resp, err := s.client.Grant(ctx, lease)
	if err != nil {
		return err
	}

	// 设置key并绑定租约
	_, err = s.client.Put(ctx, s.key, strconv.Itoa(s.weight), clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// 设置KeepAlive自动续租，注意不能使用上面的ctx，KeepAlive续租的context不能关闭
	keepAliveChan, err := s.client.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	s.leaseID = resp.ID
	s.keepAliveChan = keepAliveChan
	log.Printf("registry put key:%s weight:%d success\n", s.key, s.weight)
	return nil
}

// ListenKeepAlive 监听自动续租
func (s *ServiceRegister) ListenKeepAlive() {
	for ka := range s.keepAliveChan {
		log.Printf("续租成功: %v, ttl: %d", ka, ka.TTL)
	}
	log.Println("关闭续租")
}

// Close 注销服务
func (s *ServiceRegister) Close() error {
	// 撤销租约
	if _, err := s.client.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}

	log.Println("撤销租约")
	return nil
}
