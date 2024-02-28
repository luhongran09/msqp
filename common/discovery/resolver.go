package discovery

import (
	"common/config"
	"common/logs"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type Resolver struct {
	schema      string
	etcdCli     *clientv3.Client
	closeCh     chan struct{}
	DialTimeout int
	conf        config.EtcdConf
	srvAddrList []resolver.Address
	cc          resolver.ClientConn
	key         string
	watchCh     clientv3.WatchChan
}

// Build 当grpc.Dial 的时候 就会同步调用此方法
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	//获取到调用的key （user/v1）连接etcd 获取其value
	//1.连接etcd
	var err error
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.conf.Addrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		logs.Fatal("connect etcd failed,err: %v", err)
	}
	//2.根据key获取value
	r.closeCh = make(chan struct{})
	r.key = target.URL.Host
	r.sync()
	return nil, nil
}

func (r Resolver) Scheme() string {

}
func (r Resolver) sync() error {
	ctx, cannel := context.WithTimeout(context.Background(), time.Duration(r.conf.RWTimeout)*time.Second)
	defer cannel()
	res, err := r.etcdCli.Get(ctx, r.key, clientv3.WithPrefix())
	if err != nil {
		logs.Error("get etcd register service failed ,name= %s,err :%v", r.key, err)
		return err
	}
	r.srvAddrList = []resolver.Address{}
	for _, v := range res.Kvs {
		server, err := ParseValue(v.Value)
		if err != nil {
			logs.Error("parse etcd register service failed,name=%s,err:%v", r.key, err)
			continue
		}
		r.srvAddrList = append(r.srvAddrList, resolver.Address{
			Addr:       server.Addr,
			Attributes: attributes.New("weight", server.Weight),
		})
	}
	r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList})
	if err != nil {
		logs.Error("updateState etcd register service failed,name=%s,err:%v", r.key, err)
		return err
	}
	return nil
}

func NewResolver(conf config.EtcdConf) *Resolver {
	return &Resolver{}
}
