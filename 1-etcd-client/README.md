## etcd

### 简介

[etcd](https://github.com/etcd-io/etcd)是开源的、高可用的分布式key-value存储系统，可用于配置共享、服务的注册和发现，主要特点如下：

* 简单：定义清晰、面向用户的API（gRPC）

* 安全：可选的客户端TLS证书自动认证

* 快速：支持每秒10000次写入

* 可靠：基于Raft算法确保强一致性


### etcd与redis对比

etcd和redis都支持kv键值存储，也支持分布式特性，redis支持的数据格式更加丰富，但是他们两个定位和应用场景不一样，主要差异如下：

* redis在分布式环境下不是强一致性的，可能会丢失数据，或者读取不到最新数据

* redis的数据变化监听机制没有etcd完善

* etcd强一致性保证数据可靠性，导致性能上要低于redis


### etcdctl

```
etcdctl version
etcdctl put /test/foo "bar"
etcdctl get /test/foo
etcdctl del /test/foo

# watch监听
etcdctl watch /test/foo
etcdctl watch --prefix /test/foo

# 新建一个过期时间为60s的租约
$ etcdctl lease grant 60
lease xxx granted with TTL(60s)

# 查看租约列表
etcdctl lease list

# 绑定租约
etcdctl put /test/foo "bar" --lease=xxx

# 撤销租约，将删除所有绑定的key
etcdctl lease revoke xxx

# 查看租约信息
etcdctl lease timetolive xxx --keys

ETCDCTL_API=3 etcdctl --endpoints 127.0.0.1:2379 endpoint status --write-out="table"
etcdctl cluster-health
ectdctl member list
```