package elog

import (
	"github.com/jingleizhang/elog"
)

func init() {
	statLogger = elog.NewSingleton(newStatLogger)
}

var statLogger elog.Singleton

func newStatLogger() (interface{}, error) {
	logger := elog.NewELog("stat_logger", elog.LOG_SHIFT_BY_DAY, 2*1024*1024, elog.LOG_FATAL|elog.LOG_ERROR|elog.LOG_INFO|elog.LOG_DEBUG)
	logger.SetKeptInFile(true)
	return logger, nil
}

func Logger() *elog.ELog {
	logger, err := statLogger.Get()
	if logger != nil && err == nil {
		return logger.(*elog.ELog)
	}
	return nil
}
