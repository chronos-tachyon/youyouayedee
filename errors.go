package youyouayedee

import (
	"bytes"
	"fmt"
	"strconv"
)

// ErrClockNotFound indicates that the ClockStorage Load method was unable to
// provide a (last known timestamp, last known counter value) tuple for the
// given Node.
type ErrClockNotFound struct{}

func (ErrClockNotFound) Error() string {
	return "clock data not found"
}

var _ error = ErrClockNotFound{}

// ErrLockNotSupported indicates that file locking is not supported on the
// current OS platform.
type ErrLockNotSupported struct{}

func (ErrLockNotSupported) Error() string {
	return "file locking not supported"
}

var _ error = ErrLockNotSupported{}

// ErrVersionNotSupported indicates that NewGenerator does not know how to
// generate UUIDs of the given Version.
type ErrVersionNotSupported struct {
	Version Version
}

func (err ErrVersionNotSupported) Error() string {
	return fmt.Sprintf("%v UUIDs are not supported", err.Version)
}

var _ error = ErrVersionNotSupported{}

// ErrVersionMismatch indicates that a Generator constructor is not implemented
// for UUIDs of the given Version.
type ErrVersionMismatch struct {
	Requested Version
	Expected  []Version
}

func (err ErrVersionMismatch) Error() string {
	buf := make([]byte, 0, 64)
	buf = append(buf, err.Requested.String()...)
	buf = append(buf, " UUIDs are not supported by this generator; only "...)
	expected := err.Expected
	expectedLen := len(expected)
	expectedLast := expectedLen - 1
	for index := 0; index < expectedLen; index++ {
		if index != 0 {
			if index == expectedLast {
				buf = append(buf, ", and "...)
			} else {
				buf = append(buf, ", "...)
			}
		}
		buf = append(buf, expected[index].String()...)
	}
	buf = append(buf, " UUIDs are supported"...)
	return string(buf)
}

var _ error = ErrVersionMismatch{}

// ErrHashFactoryIsNil indicates that a Generator requires a hash.Hash factory
// callback.
type ErrHashFactoryIsNil struct {
	Version Version
}

func (err ErrHashFactoryIsNil) Error() string {
	return fmt.Sprintf("this generator for %v UUIDs requires a hash.Hash factory callback, but factory is nil", err.Version)
}

var _ error = ErrHashFactoryIsNil{}

// ErrNamespaceNotValid indicates that a Generator requires a valid namespace
// UUID.
type ErrNamespaceNotValid struct {
	Version   Version
	Namespace UUID
}

func (err ErrNamespaceNotValid) Error() string {
	return fmt.Sprintf("this generator for %v UUIDs requires a valid namespace UUID, but Namespace %v is not valid", err.Version, err.Namespace)
}

var _ error = ErrNamespaceNotValid{}

// ErrMethodNotSupported indicates that the called Generator method is not
// supported by the implementation.
type ErrMethodNotSupported struct {
	Method Method
}

func (err ErrMethodNotSupported) Error() string {
	var buf bytes.Buffer
	buf.Grow(128)
	buf.WriteString("generator does not implement method ")
	buf.WriteString(err.Method.String())
	return buf.String()
}

var _ error = ErrMethodNotSupported{}

// ErrOperationFailed indicates that a required step failed while initializing
// a Generator or generating a UUID.
type ErrOperationFailed struct {
	Operation Operation
	Err       error
}

func (err ErrOperationFailed) Error() string {
	var buf bytes.Buffer
	buf.Grow(128)
	buf.WriteString(err.Operation.String())
	buf.WriteString(": ")
	buf.WriteString(err.Err.Error())
	return buf.String()
}

func (err ErrOperationFailed) Unwrap() error {
	return err.Err
}

var _ error = ErrOperationFailed{}

// ErrParseFailed indicates that the input string could not be parsed as a
// UUID.
type ErrParseFailed struct {
	Input      []byte
	Problem    ParseProblem
	Args       []interface{}
	Index      uint
	ExpectByte byte
	ActualByte byte
}

func (err ErrParseFailed) Error() string {
	inputIsSafe := true
	inputLen := uint(len(err.Input))
	for ii := uint(0); ii < inputLen; ii++ {
		ch := err.Input[ii]
		if ch < 0x20 || ch >= 0x7f {
			inputIsSafe = false
			break
		}
	}

	buf := make([]byte, 0, 128)
	buf = append(buf, "failed to parse "...)
	if inputIsSafe && inputLen != 16 {
		buf = strconv.AppendQuote(buf, string(err.Input))
	} else {
		for ii := uint(0); ii < inputLen; ii++ {
			if ii != 0 {
				buf = append(buf, ':')
			}
			buf = appendHexByte(buf, err.Input[ii])
		}
	}
	buf = append(buf, " as UUID: "...)
	var formatted string
	data := err.Problem.Data()
	if data.Format == "" {
		buf = append(buf, data.Name...)
		buf = append(buf, "; "...)
		formatted = fmt.Sprintf("%#v", err.Args)
	} else {
		formatted = fmt.Sprintf(data.Format, err.Args...)
	}
	buf = append(buf, formatted...)
	return string(buf)
}

var _ error = ErrParseFailed{}
