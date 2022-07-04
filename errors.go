package youyouayedee

import (
	"bytes"
	"fmt"
	"strconv"
)

// ClockStorageUnavailableError indicates that the ClockStorage Load method was
// unable to provide a (last known timestamp, last known counter value) tuple
// for the given Node.
type ClockStorageUnavailableError struct{}

func (ClockStorageUnavailableError) Error() string {
	return "ClockStorage is not available"
}

var _ error = ClockStorageUnavailableError{}

// UnsupportedVersionError indicates that NewGenerator does not know how to
// generate UUIDs of the given Version.
type UnsupportedVersionError struct {
	Version Version
}

func (err UnsupportedVersionError) Error() string {
	return fmt.Sprintf("%v UUIDs are not supported", err.Version)
}

var _ error = UnsupportedVersionError{}

// MismatchedVersionError indicates that a Generator constructor is not
// implemented for UUIDs of the given Version.
type MismatchedVersionError struct {
	Requested Version
	Expected  []Version
}

func (err MismatchedVersionError) Error() string {
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

var _ error = MismatchedVersionError{}

// NilHashFactoryError indicates that a Generator requires a hash.Hash factory
// callback.
type NilHashFactoryError struct {
	Version Version
}

func (err NilHashFactoryError) Error() string {
	return fmt.Sprintf("this generator for %v UUIDs requires a hash.Hash factory callback, but factory is nil", err.Version)
}

var _ error = NilHashFactoryError{}

// InvalidNamespaceError indicates that a Generator requires a valid namespace
// UUID.
type InvalidNamespaceError struct {
	Version   Version
	Namespace UUID
}

func (err InvalidNamespaceError) Error() string {
	return fmt.Sprintf("this generator for %v UUIDs requires a valid namespace UUID, but Namespace %v is not valid", err.Version, err.Namespace)
}

var _ error = InvalidNamespaceError{}

// MethodNotSupportedError indicates that the called Generator method is not
// supported by the implementation.
type MethodNotSupportedError struct {
	Method Method
}

func (err MethodNotSupportedError) Error() string {
	var buf bytes.Buffer
	buf.Grow(128)
	buf.WriteString("generator does not implement method ")
	buf.WriteString(err.Method.String())
	return buf.String()
}

var _ error = MethodNotSupportedError{}

// FailedOperationError indicates that a required step failed while
// initializing a Generator or generating a UUID.
type FailedOperationError struct {
	Operation Operation
	Err       error
}

func (err FailedOperationError) Error() string {
	var buf bytes.Buffer
	buf.Grow(128)
	buf.WriteString(err.Operation.String())
	buf.WriteString(": ")
	buf.WriteString(err.Err.Error())
	return buf.String()
}

func (err FailedOperationError) Unwrap() error {
	return err.Err
}

var _ error = FailedOperationError{}

// ParseError indicates that the input string could not be parsed as a UUID.
type ParseError struct {
	Input      []byte
	Problem    ParseProblem
	Args       []interface{}
	Index      uint
	ExpectByte byte
	ActualByte byte
}

func (err ParseError) Error() string {
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

var _ error = ParseError{}

// IOError indicates that an I/O error or OS system call error occurred.
type IOError struct {
	Err error
}

func (err IOError) Error() string {
	return fmt.Sprintf("I/O error: %s", err.Err.Error())
}

func (err IOError) Unwrap() error {
	return err.Err
}

var _ error = IOError{}

func mkargs(args ...interface{}) []interface{} {
	return args
}
