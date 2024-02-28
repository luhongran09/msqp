package rpc

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"user/pb"
)

var (
	UserClient pb.UserServiceClient
)

func Init() {
	//etcd解析器 就可以grpc连接的时候 进行触发 通过提供的addr地址 去etcd中进行查找
	r := discovery.NewResolver()
	resolver.Register(r)
	domain := config.Conf.Domain["user"]
	initClient()
}

func InitClient(name string, loadBalance bool, client interface{}) {
	//找服务的地址
	addr := fmt.Sprintf("etcd:///%s", name)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials())}
	if loadBalance {
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`"{LoadBalancingPolicy}":`)))
	}
	conn, err := grpc.DialContext(context.TODO(), addr)
	if err != nil {
		logs.Fatal("rpc connect etcd err:%v", err)
	}
	switch c := client.(type) {
	case *pb.UserServiceClient:
		*c = pb.NewUserServiceClient(conn)
	}
 
}
