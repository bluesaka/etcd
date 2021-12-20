package main

import (
	"context"
	"log"
	"sync"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
)

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	cli        *clientv3.Client //etcd client
	serverList sync.Map
}

// NewServiceDiscovery  新建服务发现
func NewServiceDiscovery(endpoints []string) *ServiceDiscovery {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		log.Fatal(err)
	}

	return &ServiceDiscovery{
		cli: cli,
	}
}

// Watch 监听
func (s *ServiceDiscovery) Watch(prefix string) error {
	resp, err := s.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	for _, ev := range resp.Kvs {
		s.SetServiceList(string(ev.Key), string(ev.Value))
	}

	go s.watcher(prefix)
	return nil
}

func (s *ServiceDiscovery) watcher(prefix string) {
	watchChan := s.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for w := range watchChan {
		for _, ev := range w.Events {
			switch ev.Type {
			case mvccpb.PUT: // 修改或者新增
				s.SetServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: // 删除
				s.DelServiceList(string(ev.Kv.Key))
			}
		}
	}
}

// SetServiceList 新增服务地址
func (s *ServiceDiscovery) SetServiceList(key, val string) {
	s.serverList.Store(key, val)
	log.Println("put key :", key, "val:", val)
}

// DelServiceList 删除服务地址
func (s *ServiceDiscovery) DelServiceList(key string) {
	s.serverList.Delete(key)
	log.Println("del key:", key)
}

// GetServices 获取服务地址
func (s *ServiceDiscovery) GetServiceList() []string {
	addrs := make([]string, 0, 10)
	s.serverList.Range(func(k, v interface{}) bool {
		addrs = append(addrs, v.(string))
		return true
	})
	return addrs
}

// Close 关闭服务
func (s *ServiceDiscovery) Close() error {
	return s.cli.Close()
}

func main() {
	var endpoints = []string{"localhost:2379"}
	s := NewServiceDiscovery(endpoints)
	defer s.Close()

	log.Println("Start discovery...")

	// 监听key
	s.Watch("/HTTP/")
	s.Watch("/gRPC/")

	for {
		select {
		case <-time.Tick(10 * time.Second):
			log.Println("service list:", s.GetServiceList())
		}
	}
}
