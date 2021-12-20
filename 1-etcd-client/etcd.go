package etcd

import (
	"context"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"time"
)

var endpoints = []string{"127.0.0.1:2379"}

// Base 基础使用
func Base() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.Put(ctx, "test-key", "value1")
	if err != nil {
		log.Println("etcd put failed, err:", err)
	}

	//resp, err := client.Get(ctx, "test-key", clientv3.WithPrefix())
	resp, err := client.Get(ctx, "test-key")
	if err != nil {
		log.Println("etcd get failed, err:", err)
	}
	log.Println(resp)
	for _, v := range resp.Kvs {
		log.Printf("key: %s, value: %s, version: %d\n", v.Key, v.Value, v.Version)
	}
}

// Watch 监听
func Watch() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// clientv3.WithXXX
	watchChan := client.Watch(context.Background(), "test-key")
	log.Println("etcd start watching")

	for resp := range watchChan {
		for _, v := range resp.Events {
			log.Printf("watch type: %s, key: %s, value: %s", v.Type, v.Kv.Key, v.Kv.Value)
		}
	}
}

// Lease 租约
func Lease() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}

	log.Println("connect to etcd success.")
	defer client.Close()

	// 创建一个5秒的租约
	resp, err := client.Grant(context.Background(), 5)
	if err != nil {
		log.Fatal(err)
	}

	// 5秒钟之后, /test/ 这个key就会被移除
	_, err = client.Put(context.Background(), "/lease-key-a", "lease-a", clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
	}

	r, err := client.Get(context.Background(), "/lease-key-a")
	if err != nil {
		log.Println("etcd get failed, err:", err)
	}
	log.Println(r)
	for _, v := range r.Kvs {
		log.Printf("key: %s, value: %s, version: %d\n", v.Key, v.Value, v.Version)
	}

	time.Sleep(time.Second * 6)

	r, err = client.Get(context.Background(), "/lease-key-a")
	if err != nil {
		log.Println("etcd get failed, err:", err)
	}
	log.Println(r)
	for _, v := range r.Kvs {
		log.Printf("key: %s, value: %s, version: %d\n", v.Key, v.Value, v.Version)
	}
}

// KeepAlive 保活
func KeepAlive() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}
	log.Println("connect to etcd success.")
	defer client.Close()

	resp, err := client.Grant(context.TODO(), 5)
	if err != nil {
		log.Fatal(err)
	}

	_, err = client.Put(context.TODO(), "/test-key-b", "test-b", clientv3.WithLease(resp.ID))
	if err != nil {
		log.Fatal(err)
	}

	// the key will be kept forever
	ch, err := client.KeepAlive(context.TODO(), resp.ID)
	if err != nil {
		log.Fatal(err)
	}
	for {
		ka := <-ch
		log.Printf("LeaseKeepAliveResponse: %v, leaseID: %d, ttl: %d", ka, ka.ID, ka.TTL)
	}
}

// Lock 锁
func Lock() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 使用Status来检测endpoint
	if _, err := client.Status(ctx, endpoints[0]); err != nil {
		panic(err)
	}

	log.Println("connect to etcd success.")
	defer client.Close()

	// 创建两个单独的会话用来演示锁竞争
	s1, err := concurrency.NewSession(client)
	if err != nil {
		panic(err)
	}
	defer s1.Close()
	m1 := concurrency.NewMutex(s1, "/my-lock/")

	s2, err := concurrency.NewSession(client)
	if err != nil {
		panic(err)
	}
	defer s2.Close()
	m2 := concurrency.NewMutex(s2, "/my-lock/")

	if err := m1.Lock(context.Background()); err != nil {
		log.Println("lock error:", err)
	}
	log.Println("acquired lock for s1")

	m2Locked := make(chan struct{}, 1)
	go func() {
		defer close(m2Locked)
		// 等待直到s1释放了/my-lock/的锁
		if err := m2.Lock(context.Background()); err != nil {
			log.Println("lock error:", err)
		}
	}()

	if err := m1.Unlock(context.Background()); err != nil {
		log.Println("unlock error:", err)
	}
	log.Println("released lock for s1")

	<-m2Locked
	log.Println("acquired lock for s2")
}
