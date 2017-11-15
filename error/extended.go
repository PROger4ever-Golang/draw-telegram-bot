package eepkg

import (
	"fmt"
	"io"
)

type ExtendedError struct {
	*ExtendedError
	msg      string
	cause    error
	data     interface{}
	stack    *Stack
	original error
}

func (ee *ExtendedError) Error() string {
	res := ee.msg
	if ee.cause != nil {
		res += ": " + ee.cause.Error()
	}
	return res
}

func (ee *ExtendedError) Cause() error {
	return ee.cause
}

func (ee *ExtendedError) Data() interface{} {
	return ee.data
}

func (ee *ExtendedError) StackTrace() StackTrace {
	return ee.stack.StackTrace()
}

func (ee *ExtendedError) Original() error {
	return ee.original
}

func (ee *ExtendedError) IsRoot() bool {
	return ee.cause == nil
}

func (ee *ExtendedError) GetRoot() error {
	if ee.original == nil {
		return ee
	}
	return ee.original
}

func (ee *ExtendedError) AsError() error {
	if ee == nil {
		return nil
	}
	return ee
}

func (ee *ExtendedError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, ee.msg)
			fmt.Fprintf(s, "\nData: %+v", ee.data)
			fmt.Fprint(s, "\nStack:\n")
			if ee.stack == nil {
				fmt.Fprintf(s, "nil")
			} else {
				ee.stack.Format(s, verb)
			}
			fmt.Fprintf(s, "\nCause: %+v", ee.cause)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, ee.Error())
	case 'q':
		fmt.Fprintf(s, "%q", ee.Error())
	}
}

func (ee *ExtendedError) setCause(cause error) *ExtendedError {
	ee.cause = cause

	causeEE, isEE := cause.(*ExtendedError)
	if isEE && !causeEE.IsRoot() {
		ee.original = causeEE.original
	} else {
		ee.original = cause
	}
	return ee
}

func newInternal(cause error, data interface{}, enableStack bool, msg string) *ExtendedError {
	w := &ExtendedError{
		msg:  msg,
		data: data,
	}
	if enableStack {
		w.stack = TraceStack()
	}
	return w.setCause(cause)
}

func New(data interface{}, enableStack bool, msg string) *ExtendedError {
	return newInternal(nil, data, enableStack, msg)
}

func Newf(data interface{}, enableStack bool, format string, args ...interface{}) *ExtendedError {
	msg := fmt.Sprintf(format, args...)
	return newInternal(nil, data, enableStack, msg)
}

func Wrap(cause error, data interface{}, enableStack bool, msg string) *ExtendedError {
	if cause == nil {
		return nil
	}
	return newInternal(cause, data, enableStack, msg)
}

func Wrapf(cause error, data interface{}, enableStack bool, format string, args ...interface{}) *ExtendedError {
	if cause == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return newInternal(cause, data, enableStack, msg)
}
