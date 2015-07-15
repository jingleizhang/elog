package main

import (
	"github.com/jingleizhang/elog/logger"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(2)

	elog.Logger().Error("this is logger test for error|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")

	elog.Logger().Info("this is logger test for info|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")
}
