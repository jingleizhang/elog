package elog

import (
	"github.com/jingleizhang/elog"
)

func init() {
	statLogger = base.NewSingleton(newStatLogger)
}

var statLogger elog.Singleton

func newStatLogger() (interface{}, error) {
	logger := elog.NewELog("stat_logger", baselog.LOG_SHIFT_BY_DAY, 2*1024*1024, elog.LOG_FATAL|elog.LOG_ERROR|elog.LOG_INFO)
	logger.SetKeptInFile(true)
	return logger, nil
}

func GetStatLogger() *elog.ELog {
	logger, err := statLogger.Get()
	if logger != nil && err == nil {
		return logger.(*elog.ELog)
	}
	return nil
}
