package discovery

import (
	"common/config"
	"common/logs"
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

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

// CreateLease 创建租约
// expire 祖约时间 单位秒
func (r *Register) createLease(ctx context.Context, expire int64) error {
	grant, err := r.etcdCli.Grant(ctx, expire)
	if err != nil {
		logs.Error("CreateLease failed,error :%v", err)
		return err
	}
	r.leaseId = grant.ID
	return nil
}

// BindLease 绑定租约
func (r *Register) bindLease(ctx context.Context, key, value string) error {
	_, err := r.etcdCli.Put(ctx, key, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		logs.Error("bindLease failed,error:%v", err)
		return err
	}
	return nil
}

// KeepAlive 心跳，确保服务正常
func (r *Register) keepAlive(ctx context.Context) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	resChan, err := r.etcdCli.KeepAlive(ctx, r.leaseId)
	if err != nil {
		logs.Error("bindLease failed, error:%v", err)
		return resChan, err
	}
	return resChan, nil
}

// Watcher 监听 续租 注销等
func (r *Register) watcher() {
	ticker := time.NewTicker(time.Duration(r.info.Ttl) * time.Second)
	for {
		select {
		case <-r.closeCh:
			logs.Info("stop register...")
			//注销
			if err := r.unregister(); err != nil {
				logs.Error("stop register,unregister failed,error:%v", err)
			}
			//撤销租约
			if _, err := r.etcdCli.Revoke(context.Background(), r.leaseId); err != nil {
				logs.Error("stop register,revoke failed ,error:%v", err)
			}
		case res := <-r.keepAliveCh:
			//续约
			if res == nil {
				if err := r.register(); err != nil {
					logs.Error("keepAliveCh,register failed,error:%v", err)
				}
			}
		case <-ticker.C:
			if r.keepAliveCh == nil {
				//租约到期，检查是否需要重新注册
				if err := r.register(); err != nil {
					logs.Error("ticker.c keepAliveCh==nil,register failed,error:%v", err)
				}
			}
		}
	}
}
func (r *Register) unregister() error {
	_, err := r.etcdCli.Delete(context.Background(), r.info.BuildRegisterKey())
	r.info.BuildRegisterKey()
	return err
}

func (r *Register) register() error {
	ctx, cannel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cannel()
	var err error
	if err = r.createLease(ctx, r.info.Ttl); err != nil {
		return err
	}
	if r.keepAliveCh, err = r.keepAlive(ctx); err != nil {
		return err
	}
	data, err := json.Marshal(r.info)
	if err != nil {
		logs.Error("etcd register json marshal error:%v", err)
		return err
	}
	return r.bindLease(ctx, r.info.BuildRegisterKey(), string(data))
}

func (r *Register) Register(conf config.EtcdConf) error {
	info := Server{
		Name:    conf.Register.Name,
		Addr:    conf.Register.Addr,
		Version: conf.Register.Version,
		Weight:  conf.Register.Weight,
		Ttl:     conf.Register.Ttl,
	}
	var err error
	if r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Addrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	}); err != nil {
		return err
	}
	r.info = info
	if err = r.register(); err != nil {
		return err
	}
	r.closeCh = make(chan struct{})
	go r.watcher()
	return nil
}

func (r *Register) Stop() {
	r.closeCh <- struct{}{}
}
