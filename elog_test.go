package elog

import (
	"fmt"
	"testing"
	"time"
)

var testLogger = NewSingleton(newTestLog)

func newTestLog() (interface{}, error) {
	logger := NewELog("debug", LOG_SHIFT_BY_MIN, 100, LOG_FATAL|LOG_ERROR|LOG_INFO)
	return logger, nil
}

func getTestLogger() *ELog {
	logger, err := testLogger.Get()
	if logger != nil && err == nil {
		return logger.(*ELog)
	}
	return nil
}

func TestBaselog(t *testing.T) {
	cnt := 10
	ch := make(chan int, cnt)

	for i := 1; i < cnt; i++ {
		go func(index int) {
			logger := getTestLogger()
			if logger == nil {
				t.Error("got nil logger.")
			}

			for j := 0; j < 100; j++ {
				sum := index*10000 + j
				logger.Info("i|", index, "|sum|", sum)
				time.Sleep(1 * time.Second)
			}

			ch <- index
		}(i)
	}

	for i := 1; i < cnt; i++ {
		index := <-ch
		fmt.Println("finish ", index)
	}
}
