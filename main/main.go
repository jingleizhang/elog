package main

import (
	"github.com/jingleizhang/elog/logger"
)

func main() {
	elog.Logger().VIP("this is logger test for VIP|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")

	elog.Logger().Error("this is logger test for Error|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")

	elog.Logger().Info("this is logger test for Info|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")

	elog.Logger().Debug("this is logger test for Debug|int|", 100, "|string|", "hello", "|float32|", 20.15, "|struct|etc.")
}
