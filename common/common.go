package common

import (
	"fmt"
	"os"

	ee "bitbucket.org/proger4ever/draw-telegram-bot/errors"
)

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	if x == 0 {
		return 0
	}
	return x
}

func PanicIfError(err error, msg string) {
	if err != nil {
		panic(fmt.Errorf("Error occured while %v:\n%q", msg, err))
	}
}

func TraceIfPanic(src string, data interface{}) {
	if err := recover(); err != nil {
		w := ee.Wrapf(err.(error), data, true, "Error occured in %s", src)
		fmt.Fprintf(os.Stderr, "%+v", w)
	}
}
