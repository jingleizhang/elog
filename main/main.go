package main

import (
	"github.com/jingleizhang/elog/logger"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)

	elog.GetStatLogger().DLogInfo("this is logger test|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")
}