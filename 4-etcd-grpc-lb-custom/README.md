### 自定义gRPC负载均衡策略

gRPC提供了PickerBuilder和Picker接口，让我们实现自己的负载均衡策略。

```
type PickerBuilder interface {
	Build(info PickerBuildInfo) balancer.Picker
}

PickerBuilder接口：创建subConn子连接选择器。
Build方法：返回一个选择器，将用于gRPC选择子连接。
```

```
type Picker interface {
	Pick(info PickInfo) (PickResult, error)
}

Picker接口：用于gRPC选择子连接去发送请求。
Pick方法：子连接选择
```

### 加权随机法

我们需要把服务器地址的权重添加进去，但是`resolver.Address`并没有提供权重的属性，可以把权重存储到地址的元数据`metadata`，参考`weight.go`。


### 运行

注册服务，启动server

```
$ go run server.go --port=:8001 --weight=5
2021-10-15 18:16:25.572244 I | registry put key:/grpc-lb-custom/grpclb_test2/localhost:8001 weight:5 success
2021-10-15 18:16:25.572284 I | server starting on :8001

$ go run server.go --port=:8002 --weight=1
2021-10-15 18:16:27.636077 I | registry put key:/grpc-lb-custom/grpclb_test2/localhost:8002 weight:1 success
2021-10-15 18:16:27.636121 I | server starting on :8002
```

服务发现，启动client

```
go run client.go
```

查看server端响应

```
$ go run server.go --port=:8001 --weight=5
2021-10-15 18:16:25.572244 I | registry put key:/grpc-lb-custom/grpclb_test2/localhost:8001 weight:5 success
2021-10-15 18:16:25.572284 I | server starting on :8001
2021-10-15 18:16:32.549510 I | receive:grpc2
2021-10-15 18:16:32.550118 I | receive:grpc5
2021-10-15 18:16:32.550311 I | receive:grpc6
2021-10-15 18:16:32.550465 I | receive:grpc7
2021-10-15 18:16:32.550610 I | receive:grpc8
2021-10-15 18:16:32.550734 I | receive:grpc9
2021-10-15 18:16:32.550878 I | receive:grpc10
2021-10-15 18:16:32.551029 I | receive:grpc11
2021-10-15 18:16:32.551173 I | receive:grpc12
2021-10-15 18:16:32.551365 I | receive:grpc13
2021-10-15 18:16:32.620771 I | receive:grpc15
2021-10-15 18:16:32.621182 I | receive:grpc16
2021-10-15 18:16:32.621499 I | receive:grpc17
2021-10-15 18:16:32.623142 I | receive:grpc20

$ go run server.go --port=:8002 --weight=1
2021-10-15 18:16:27.636077 I | registry put key:/grpc-lb-custom/grpclb_test2/localhost:8002 weight:1 success
2021-10-15 18:16:27.636121 I | server starting on :8002
2021-10-15 18:16:32.548987 I | receive:grpc1
2021-10-15 18:16:32.549718 I | receive:grpc3
2021-10-15 18:16:32.549907 I | receive:grpc4
2021-10-15 18:16:32.620410 I | receive:grpc14
2021-10-15 18:16:32.621755 I | receive:grpc18
2021-10-15 18:16:32.622955 I | receive:grpc19
```

可以看到8001端口服务的权重大，被调用的次数更多

