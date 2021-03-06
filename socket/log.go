package socket

import (
	"fmt"
	"ygo/src/libs/log"
)

const (
	LOG_CONFIG_FILE = "socket_log.conf"
)

var (
	gLog *logger.Logger
)

func init() {
	initLogger()
}

func initLogger() {
	conf := logger.NewLogConfig(LOG_CONFIG_FILE)
	err := conf.LoadConfig()
	if err != nil {
		fmt.Errorf("load log conf file failed: %v", err)
	}
	gLog = logger.NewLogger(conf)
	gLog.Infof("init logger %s done, %v", "ddd", conf)
}

func logDebug(args ...interface{}) {
	gLog.Debug(args...)
}

func logDebugf(format string, args ...interface{}) {
	gLog.Debugf(format, args...)
}

func logInfo(args ...interface{}) {
	gLog.Info(args...)
}

func logInfof(format string, args ...interface{}) {
	gLog.Infof(format, args...)
}

func logWarning(args ...interface{}) {
	gLog.Warning(args...)
}

func logWarningf(format string, args ...interface{}) {
	gLog.Warningf(format, args...)
}

func logError(args ...interface{}) {
	gLog.Error(args...)
}

func logErrorf(format string, args ...interface{}) {
	gLog.Errorf(format, args...)
}

func logFatal(args ...interface{}) {
	gLog.Fatal(args...)
}

func logFatalf(format string, args ...interface{}) {
	gLog.Fatalf(format, args...)
}