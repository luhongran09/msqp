package app

import (
	"common/config"
	"common/logs"
	"context"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run 启动程序 启动grpc服务 启动http服务 启用日志 启用数据库
func Run(ctx context.Context) error {
	server := grpc.NewServer()
	go func() {
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("user grpc server listen err:%v", err)
		}
		logs.Info("user grpc server started listen on %s", config.Conf.Grpc.Addr)
		if err = server.Serve(lis); err != nil {
			logs.Fatal("run user grpc server failed,err:%v", err)
		}
	}()
	c := make(chan os.Signal, 1)
	stop := func() {
		server.Stop()
		time.Sleep(3 * time.Second) //给3秒时间停止必要的服务
	}
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-ctx.Done():
			return nil
		case s := <-c:
			logs.Warn("get s signal %s", s.String())
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				stop()
				logs.Warn("user grpc server exit")
				return nil
			case syscall.SIGHUP:
			default:
				return nil
			}
		}
	}
}
