package app

import (
	"common/config"
	"common/logs"
	"context"
	"fmt"
	"gate/router"
	"os"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context) error {
	logs.InitLog(config.Conf.AppName)
	go func() {
		r, err := router.RegisterRouter()
		if err != nil {
			logs.Fatal("user module gin register router error : %v", err)
		}
		err = r.Run(fmt.Sprintf(":%d", config.Conf.HttpPort))
		if err != nil {
			logs.Fatal("gate gin run err: %v", err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case <-ctx.Done():
			return nil
		case s := <-c:
			logs.Warn("get a signal %s", s.String())
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				logs.Warn("gate  exit")
				return nil
			case syscall.SIGHUP:
			default:
				return nil
			}
		}
	}
}
