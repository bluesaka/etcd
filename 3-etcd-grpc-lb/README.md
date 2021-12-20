## gRPC负载均衡（客户端负载均衡）

### gRPC负载均衡
gRPC官方文档提供了关于gRPC负载均衡方案[Load Balancing in gRPC](https://github.com/grpc/grpc/blob/master/doc/load-balancing.md)
提供了 `pick_first` （默认：选择第一个） 和 `round_robin` （轮询） 两种负载均衡方式

### 运行

- 启动并注册三个服务

```
$ go run server.go --port=:8001
2021-10-13 15:32:35.777148 I | registry put key:/grpc-lb/grpclb_test1/localhost:8001 val:localhost:8001 success
2021-10-13 15:32:35.777181 I | server starting on :8001

$ go run server.go --port=:8002
2021-10-13 15:32:47.110217 I | registry put key:/grpc-lb/grpclb_test1/localhost:8002 val:localhost:8002 success
2021-10-13 15:32:47.110247 I | server starting on :8002

$ go run server.go --port=:8003
2021-10-13 15:32:53.093678 I | registry put key:/grpc-lb/grpclb_test1/localhost:8003 val:localhost:8003 success
2021-10-13 15:32:53.093714 I | server starting on :8003
```

- 然后客户端进行调用

```
$ go run client.go
2021-10-13 15:33:16.485866 I | discovery put key:/grpc-lb/grpclb_test1/localhost:8001 val:localhost:8001
2021-10-13 15:33:16.485929 I | discovery put key:/grpc-lb/grpclb_test1/localhost:8002 val:localhost:8002
2021-10-13 15:33:16.485963 I | discovery put key:/grpc-lb/grpclb_test1/localhost:8003 val:localhost:8003
2021-10-13 15:33:16.486071 I | discovery watch
2021-10-13 15:33:16.490392 I | code:200 value:"hello grpc1" 
2021-10-13 15:33:16.490870 I | code:200 value:"hello grpc2" 
2021-10-13 15:33:16.491359 I | code:200 value:"hello grpc3" 
2021-10-13 15:33:16.491596 I | code:200 value:"hello grpc4" 
2021-10-13 15:33:16.491835 I | code:200 value:"hello grpc5" 
2021-10-13 15:33:16.492067 I | code:200 value:"hello grpc6" 
2021-10-13 15:33:16.492269 I | code:200 value:"hello grpc7" 
2021-10-13 15:33:16.492515 I | code:200 value:"hello grpc8" 
2021-10-13 15:33:16.492735 I | code:200 value:"hello grpc9" 
2021-10-13 15:33:16.492948 I | code:200 value:"hello grpc10" 
2021-10-13 15:33:16.493165 I | code:200 value:"hello grpc11" 
2021-10-13 15:33:16.493371 I | code:200 value:"hello grpc12" 
2021-10-13 15:33:16.493620 I | code:200 value:"hello grpc13" 
2021-10-13 15:33:16.493853 I | code:200 value:"hello grpc14" 
2021-10-13 15:33:16.494047 I | code:200 value:"hello grpc15" 
2021-10-13 15:33:16.494257 I | code:200 value:"hello grpc16" 
2021-10-13 15:33:16.494434 I | code:200 value:"hello grpc17" 
2021-10-13 15:33:16.494658 I | code:200 value:"hello grpc18" 
2021-10-13 15:33:16.494831 I | code:200 value:"hello grpc19" 
2021-10-13 15:33:16.495060 I | code:200 value:"hello grpc20" 
2021-10-13 15:33:16.495069 I | discovery close
```

- 看服务端接收到的请求

```
$ go run server.go --port=:8001
2021-10-13 15:32:35.777148 I | registry put key:/grpc-lb/grpclb_test1/localhost:8001 val:localhost:8001 success
2021-10-13 15:32:35.777181 I | server starting on :8001
2021-10-13 15:33:16.491206 I | receive:grpc3
2021-10-13 15:33:16.491921 I | receive:grpc6
2021-10-13 15:33:16.492629 I | receive:grpc9
2021-10-13 15:33:16.493277 I | receive:grpc12
2021-10-13 15:33:16.493960 I | receive:grpc15
2021-10-13 15:33:16.494554 I | receive:grpc18

$ go run server.go --port=:8002
2021-10-13 15:32:47.110217 I | registry put key:/grpc-lb/grpclb_test1/localhost:8002 val:localhost:8002 success
2021-10-13 15:32:47.110247 I | server starting on :8002
2021-10-13 15:33:16.490161 I | receive:grpc1
2021-10-13 15:33:16.491447 I | receive:grpc4
2021-10-13 15:33:16.492168 I | receive:grpc7
2021-10-13 15:33:16.492848 I | receive:grpc10
2021-10-13 15:33:16.493477 I | receive:grpc13
2021-10-13 15:33:16.494165 I | receive:grpc16
2021-10-13 15:33:16.494744 I | receive:grpc19

$ go run server.go --port=:8003
2021-10-13 15:32:53.093678 I | registry put key:/grpc-lb/grpclb_test1/localhost:8003 val:localhost:8003 success
2021-10-13 15:32:53.093714 I | server starting on :8003
2021-10-13 15:33:16.490721 I | receive:grpc2
2021-10-13 15:33:16.491712 I | receive:grpc5
2021-10-13 15:33:16.492390 I | receive:grpc8
2021-10-13 15:33:16.493069 I | receive:grpc11
2021-10-13 15:33:16.493722 I | receive:grpc14
2021-10-13 15:33:16.494341 I | receive:grpc17
2021-10-13 15:33:16.494942 I | receive:grpc20

可以看到grpc的轮询负载均衡策略正常
```