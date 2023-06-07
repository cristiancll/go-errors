package errs

import (
	"bytes"
	"fmt"
	"runtime"
)

type ErrorCode int32

type Error struct {
	Code     ErrorCode
	Message  string
	Metadata []any
	frame    *runtime.Frame
	source   *Error
	wrapper  *Error
}

var _ error = &Error{}

func (e *Error) Error() string {
	b := new(bytes.Buffer)
	e.printStack(b)
	return b.String()
}

func errorFrame() *runtime.Frame {
	var stack [64]uintptr
	const skip = 3
	n := runtime.Callers(skip, stack[:])
	frames := runtime.CallersFrames(stack[:n])
	f, _ := frames.Next()
	return &f
}

func New(err error, code ErrorCode, metadata ...any) *Error {
	if err == nil {
		return nil
	}
	e := &Error{
		Code:     code,
		Message:  err.Error(),
		Metadata: metadata,
		frame:    errorFrame(),
		wrapper:  nil,
	}
	e.source = e
	return e
}

func Is(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*Error)
	return ok
}

func Wrap(err error, message string, metadata ...any) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return &Error{
			Message:  message,
			Metadata: metadata,
			frame:    errorFrame(),
			source:   e.source,
			wrapper:  e,
		}
	}
	return New(err, 0, metadata...)
}

func (e *Error) printStack(b *bytes.Buffer) {
	if e == nil {
		return
	}
	if e.wrapper != nil {
		e.wrapper.printStack(b)
	}
	isSource := e.source == e
	var errorFormat string
	var metaFormat string
	if isSource {
		errorFormat = "%v:%d: %v | %s\n"
		metaFormat = "%+v\n\t"
	} else {
		errorFormat = "%v:%d: %v | %s\n\t"
		metaFormat = "%+v\n\t"
	}
	line := fmt.Sprintf(errorFormat, e.frame.File, e.frame.Line, e.frame.Function, e.Message)
	b.WriteString(line)
	if len(e.Metadata) > 0 {
		b.WriteString(fmt.Sprintf(metaFormat, e.Metadata))
	}
}
