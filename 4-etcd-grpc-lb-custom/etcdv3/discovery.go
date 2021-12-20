package etcdv3

import (
	"context"
	"etcd-example/4-3-etcd-grpc-lb-custom/balancer/weight"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

const schema = "grpc-lb-custom"

// ServiceDiscovery 服务发现
type ServiceDiscovery struct {
	client     *clientv3.Client //etcd client
	cc         resolver.ClientConn
	serverList sync.Map //服务列表
	prefix     string   //监视的前缀
}

// NewServiceDiscovery  新建服务发现
func NewServiceDiscovery(endpoints []string) resolver.Builder {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	return &ServiceDiscovery{
		client: client,
	}
}

// Build 为给定目标创建一个新的`resolver`，当调用`grpc.Dial()`时执行
func (s *ServiceDiscovery) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	s.cc = cc
	s.prefix = "/" + target.Scheme + "/" + target.Endpoint + "/"
	// 根据前缀获取现有的key
	resp, err := s.client.Get(context.Background(), s.prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, ev := range resp.Kvs {
		s.setServiceList(string(ev.Key), string(ev.Value))
	}

	s.cc.UpdateState(resolver.State{Addresses: s.getServiceList()})
	go s.watch()

	return s, nil
}

// Scheme returns schema
func (s *ServiceDiscovery) Scheme() string {
	return schema
}

// ResolveNow 监视目标更新
func (s *ServiceDiscovery) ResolveNow(options resolver.ResolveNowOptions) {
	log.Println("discovery ResolveNow")
}

// Close 关闭
func (s *ServiceDiscovery) Close() {
	log.Println("discovery close")
	s.client.Close()
}

// watch 监听
func (s *ServiceDiscovery) watch() {
	log.Println("discovery watch")
	watchChan := s.client.Watch(context.Background(), s.prefix, clientv3.WithPrefix())
	for wresp := range watchChan {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //新增或修改
				log.Println("put")
				s.setServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: //删除
				log.Println("delete")
				s.deleteServiceList(string(ev.Kv.Key))
			}
		}
	}
}

// setServiceList 新增服务地址
func (s *ServiceDiscovery) setServiceList(key, val string) {
	// 获取服务地址
	addr := resolver.Address{Addr: strings.TrimPrefix(key, s.prefix)}

	// 非数字value，权重默认为1
	w, err := strconv.Atoi(val)
	if err != nil {
		w = 1
	}
	// 把服务地址权重存储到resolver.Address的元数据中
	addr = weight.SetAddrInfo(addr, weight.AddrInfo{Weight: w})
	s.serverList.Store(key, addr)
	s.cc.UpdateState(resolver.State{Addresses: s.getServiceList()})
	log.Printf("discovery put key:%s val:%s\n", key, val)
}

// deleteServiceList 删除服务地址
func (s *ServiceDiscovery) deleteServiceList(key string) {
	s.serverList.Delete(key)
	s.cc.UpdateState(resolver.State{Addresses: s.getServiceList()})
	log.Println("discovery delete key:", key)
}

// getServiceList 获取服务地址
func (s *ServiceDiscovery) getServiceList() []resolver.Address {
	addrs := make([]resolver.Address, 0, 10)
	s.serverList.Range(func(k, v interface{}) bool {
		addrs = append(addrs, v.(resolver.Address))
		return true
	})
	return addrs
}
