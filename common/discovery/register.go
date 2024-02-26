package discovery

import clientv3 "go.etcd.io/etcd/client/v3"

// Register 将grpc注册到etcd
// 原理 创建一个租约 将grpc服务信息注册到etcd并且绑定租约
// 如果过了租约时间，etcd会删除存储的信息
// 可以实现心跳，完成续租，如果etcd没有则重新注册
type Register struct {
	etcdCli     *clientv3.Client                        //etcd 连接
	leaseId     clientv3.LeaseID                        //租约id
	DialTimeout int                                     //超时时间 秒
	ttl         int64                                   //租约时间秒
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse //心跳channel
	info        Server                                  //注册的服务信息
	closeCh     chan struct{}
}

//CreateLease 创建租约
//expire 祖约时间 单位秒
