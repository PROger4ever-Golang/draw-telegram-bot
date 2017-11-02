package common

import (
	"fmt"
	"os"

	ee "bitbucket.org/proger4ever/draw-telegram-bot/errors"
)

const traceFormat = `Panic recovered in %s.
%+v

The data occured the panic: %#v
`

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

func RepairIfError(msg string, data interface{}) {
	if err := recover(); err != nil {
		fmt.Fprintf(os.Stderr, "Recovered while %s:\n%+v\nThe data occured the panic: %v\n", msg, err, data)
	}
}

func TraceIfPanic(src string, data interface{}) {
	if err := recover(); err != nil {
		w := ee.Wrapf(err.(error), data, true, "Error occured in %s", src)
		fmt.Fprintf(os.Stderr, "%+v", w)
	}
}
