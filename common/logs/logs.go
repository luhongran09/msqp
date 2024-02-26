package logs

import (
	"common/config"
	"github.com/charmbracelet/log"
	"os"
	"time"
)

var logger *log.Logger

func InitLog(appName string) {
	logger = log.New(os.Stderr)
	if config.Conf.Log.Level == "DEBUG" {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	logger.SetPrefix(appName)           //设置日志的前缀
	logger.SetReportTimestamp(true)     //设置日志的时间戳
	logger.SetTimeFormat(time.DateTime) //设置日志的时间戳格式为“2006-01-02 15：04：05”
}
func Warn(format string, values ...any) {
	if len(values) == 0 {
		logger.Warn(format)
	} else {
		logger.Warnf(format, values...)
	}
}
func Error(format string, values ...any) {
	if len(values) == 0 {
		logger.Error(format)
	} else {
		logger.Errorf(format, values...)
	}
}
func Fatal(format string, values ...any) {
	if len(values) == 0 {
		logger.Fatal(format)
	} else {
		logger.Fatalf(format, values...)
	}
}
func Info(format string, values ...any) {
	if len(values) == 0 {
		logger.Info(format)
	} else {
		logger.Infof(format, values...)
	}
}
