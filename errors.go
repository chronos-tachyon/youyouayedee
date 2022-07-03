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

// Method enumerates the Generator methods which do not need to be implemented.
type Method uint

const (
	MethodNewUUID Method = iota
	MethodNewHashUUID
)

// MethodData holds data about a specific value of Method.
type MethodData struct {
	GoName string
	Name   string
}

var methodDataArray = [...]MethodData{
	{
		GoName: "MethodNewUUID",
		Name:   "NewUUID",
	},
	{
		GoName: "MethodNewHashUUID",
		Name:   "NewHashUUID",
	},
}

func (enum Method) Data() MethodData {
	p := uint(enum)
	q := uint(len(methodDataArray))
	if p < q {
		return methodDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.Method(%d)", p)
	name := fmt.Sprintf("<unspecified youyouayedee.Method constant %d>", p)
	return MethodData{GoName: goName, Name: name}
}

func (enum Method) GoString() string {
	return enum.Data().GoName
}

func (enum Method) String() string {
	return enum.Data().Name
}

var (
	_ fmt.GoStringer = Method(0)
	_ fmt.Stringer   = Method(0)
)

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

// Operation enumerates the operations which can fail while initializing a
// Generator or generating a UUID.
type Operation uint

const (
	GenerateNodeOp Operation = iota
	ClockStorageLoadOp
	ClockStorageStoreOp
	InitializeBlakeHashOp
)

// OperationData holds data about a specific value of Operation.
type OperationData struct {
	GoName string
	Name   string
}

var operationDataArray = [...]OperationData{
	{
		GoName: "GenerateNodeOp",
		Name:   "failed to generate node identifier",
	},
	{
		GoName: "ClockStorageLoadOp",
		Name:   "failed to obtain initial clock sequence value from persistent storage",
	},
	{
		GoName: "ClockStorageStoreOp",
		Name:   "failed to store clock sequence value to persistent storage",
	},
	{
		GoName: "InitializeBlakeHashOp",
		Name:   "failed to initialize BLAKE2B hash algorithm",
	},
}

func (enum Operation) Data() OperationData {
	p := uint(enum)
	q := uint(len(operationDataArray))
	if p < q {
		return operationDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.Operation(%d)", p)
	name := fmt.Sprintf("<unspecified youyouayedee.Operation constant %d>", p)
	return OperationData{GoName: goName, Name: name}
}

func (enum Operation) GoString() string {
	return enum.Data().GoName
}

func (enum Operation) String() string {
	return enum.Data().Name
}

var (
	_ fmt.GoStringer = Operation(0)
	_ fmt.Stringer   = Operation(0)
)

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

// ParseProblem enumerates the types of problems which can be encountered while
// parsing strings as UUIDs.
type ParseProblem uint

const (
	UnexpectedCharacter ParseProblem = iota
	WrongVariant
	WrongLength
	WrongBinaryLength
)

// ParseProblemData holds data about a specific value of ParseProblem.
type ParseProblemData struct {
	GoName string
	Name   string
	Format string
}

var parseProblemDataArray = [...]ParseProblemData{
	{
		GoName: "UnexpectedCharacter",
		Name:   "unexpected character",
		Format: "unexpected character at index %d",
	},
	{
		GoName: "WrongVariant",
		Name:   "wrong UUID variant",
		Format: "unexpected value %02x for UUID variant byte; should start with '8', '9', 'a', or 'b'",
	},
	{
		GoName: "WrongLength",
		Name:   "wrong input length",
		Format: "unexpected input length %d; should be 0, 32, 36, 38, or 41",
	},
	{
		GoName: "WrongBinaryLength",
		Name:   "wrong binary data input length",
		Format: "unexpected input length %d for binary data; should be 0, 16, 32, 36, 38, or 41",
	},
}

func (enum ParseProblem) Data() ParseProblemData {
	p := uint(enum)
	q := uint(len(parseProblemDataArray))
	if p < q {
		return parseProblemDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.ParseProblem(%d)", p)
	name := fmt.Sprintf("<unspecified youyouayedee.ParseProblem constant %d>", p)
	format := ""
	return ParseProblemData{GoName: goName, Name: name, Format: format}
}

func (enum ParseProblem) GoString() string {
	return enum.Data().GoName
}

func (enum ParseProblem) String() string {
	return enum.Data().Name
}

var (
	_ fmt.GoStringer = ParseProblem(0)
	_ fmt.Stringer   = ParseProblem(0)
)

// ParseError indicates that the input string could not be parsed as a UUID.
type ParseError struct {
	Input       string
	Problem     ParseProblem
	Args        []interface{}
	Index       uint
	Length      uint
	VariantByte byte
}

func (err ParseError) Error() string {
	data := err.Problem.Data()

	var buf bytes.Buffer
	buf.Grow(128)
	buf.WriteString("failed to parse ")
	buf.WriteString(strconv.Quote(err.Input))
	buf.WriteString(" as UUID: ")
	if data.Format == "" {
		buf.WriteString(data.Name)
	} else {
		fmt.Fprintf(&buf, data.Format, err.Args)
	}
	return buf.String()
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
