package app

import (
	"fmt"
	"os"

	"bitbucket.org/proger4ever/draw-telegram-bot/error"
)

func PanicIfExtended(err *eepkg.ExtendedError, msg string) {
	if err != nil {
		panic(fmt.Errorf("Extended error occured while %v:\n%q", msg, err))
	}
}

func TraceIfPanic(src string, data interface{}) {
	if err := recover(); err != nil {
		w := eepkg.Wrapf(err.(error), data, true, "Error occured in %s", src)
		fmt.Fprintf(os.Stderr, "%+v", w)
	}
}
