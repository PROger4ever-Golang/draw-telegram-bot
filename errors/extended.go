package errors

import (
	"fmt"
	"io"
)

type ExtendedError struct {
	error
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

func (w *ExtendedError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, w.msg)
			fmt.Fprintf(s, "\nData: %+v", w.data)
			fmt.Fprint(s, "\nStack:\n")
			if w.stack == nil {
				fmt.Fprintf(s, "nil")
			} else {
				w.stack.Format(s, verb)
			}
			fmt.Fprintf(s, "\nCause: %+v", w.cause)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, w.Error())
	case 'q':
		fmt.Fprintf(s, "%q", w.Error())
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

func New(data interface{}, enableStack bool, msg string) error {
	return newInternal(nil, data, enableStack, msg)
}

func Newf(data interface{}, enableStack bool, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return newInternal(nil, data, enableStack, msg)
}

func Wrap(cause error, data interface{}, enableStack bool, msg string) error {
	if cause == nil {
		return nil
	}
	return newInternal(cause, data, enableStack, msg)
}

func Wrapf(cause error, data interface{}, enableStack bool, format string, args ...interface{}) error {
	if cause == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	return newInternal(cause, data, enableStack, msg)
}
