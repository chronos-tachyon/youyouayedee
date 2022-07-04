package youyouayedee

import (
	"fmt"
)

// EnumData holds data about a specific enum constant.
type EnumData struct {
	GoName string
	Name   string
	Format string
}

// DCEDomain indicates the type of identifier stored in a DCE-based UUID.
type DCEDomain byte

const (
	Person DCEDomain = 0x00
	Group  DCEDomain = 0x01
	Org    DCEDomain = 0x02
)

var dceDomainDataArray = [...]EnumData{
	{GoName: "youyouayedee.Person", Name: "person"},
	{GoName: "youyouayedee.Group", Name: "group"},
	{GoName: "youyouayedee.Org", Name: "organization"},
}

func (domain DCEDomain) IsValid() bool {
	p := uint(domain)
	q := uint(len(dceDomainDataArray))
	return p < q
}

func (domain DCEDomain) Data() EnumData {
	p := uint(domain)
	q := uint(len(dceDomainDataArray))
	if p < q {
		return dceDomainDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.DCEDomain(0x%02x)", p)
	name := fmt.Sprintf("<unspecified DCE domain byte value 0x%02x>", p)
	return EnumData{GoName: goName, Name: name}
}

func (domain DCEDomain) GoString() string {
	return domain.Data().GoName
}

func (domain DCEDomain) String() string {
	return domain.Data().Name
}

var (
	_ fmt.GoStringer = DCEDomain(0)
	_ fmt.Stringer   = DCEDomain(0)
)

// Method enumerates the Generator methods which do not need to be implemented.
type Method uint

const (
	_ Method = iota
	MethodNewUUID
	MethodNewHashUUID
)

var methodDataArray = [...]EnumData{
	{
		GoName: "youyouayedee.Method(0)",
		Name:   "method not specified",
	},
	{
		GoName: "youyouayedee.MethodNewUUID",
		Name:   "NewUUID",
	},
	{
		GoName: "youyouayedee.MethodNewHashUUID",
		Name:   "NewHashUUID",
	},
}

func (enum Method) Data() EnumData {
	p := uint(enum)
	q := uint(len(methodDataArray))
	if p < q {
		return methodDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.Method(%d)", p)
	name := fmt.Sprintf("<unspecified youyouayedee.Method enum constant %d>", p)
	return EnumData{GoName: goName, Name: name}
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

// Operation enumerates the operations which can fail while initializing a
// Generator or generating a UUID.
type Operation uint

const (
	_ Operation = iota
	GenerateNodeOp
	ClockStorageLoadOp
	ClockStorageStoreOp
	InitializeBlakeHashOp
)

var operationDataArray = [...]EnumData{
	{
		GoName: "youyouayedee.Operation(0)",
		Name:   "operation not specified",
	},
	{
		GoName: "youyouayedee.GenerateNodeOp",
		Name:   "failed to generate node identifier",
	},
	{
		GoName: "youyouayedee.ClockStorageLoadOp",
		Name:   "failed to obtain initial clock sequence value from persistent storage",
	},
	{
		GoName: "youyouayedee.ClockStorageStoreOp",
		Name:   "failed to store clock sequence value to persistent storage",
	},
	{
		GoName: "youyouayedee.InitializeBlakeHashOp",
		Name:   "failed to initialize BLAKE2B hash algorithm",
	},
}

func (enum Operation) Data() EnumData {
	p := uint(enum)
	q := uint(len(operationDataArray))
	if p < q {
		return operationDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.Operation(%d)", p)
	name := fmt.Sprintf("<unspecified youyouayedee.Operation enum constant %d>", p)
	return EnumData{GoName: goName, Name: name}
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

// ParseProblem enumerates the types of problems which can be encountered while
// parsing strings as UUIDs.
type ParseProblem uint

const (
	_ ParseProblem = iota
	UnexpectedCharacter
	WrongVariant
	WrongLength
	WrongBinaryLength
)

var parseProblemDataArray = [...]EnumData{
	{
		GoName: "youyouayedee.ParseProblem(0)",
		Name:   "parse problem not specified",
	},
	{
		GoName: "youyouayedee.UnexpectedCharacter",
		Name:   "unexpected character",
		Format: "unexpected character %q at index %d; expected %s",
	},
	{
		GoName: "youyouayedee.WrongVariant",
		Name:   "wrong UUID variant",
		Format: "unexpected value %02x for UUID variant byte; should be %02x",
	},
	{
		GoName: "youyouayedee.WrongLength",
		Name:   "wrong input length",
		Format: "unexpected input length %d; should be 0, 32, 36, 38, or 41",
	},
	{
		GoName: "youyouayedee.WrongBinaryLength",
		Name:   "wrong binary data input length",
		Format: "unexpected input length %d for binary data; should be 0, 16, 32, 36, 38, or 41",
	},
}

func (enum ParseProblem) Data() EnumData {
	p := uint(enum)
	q := uint(len(parseProblemDataArray))
	if p < q {
		return parseProblemDataArray[p]
	}
	goName := fmt.Sprintf("youyouayedee.ParseProblem(%d)", p)
	name := fmt.Sprintf("<unspecified youyouayedee.ParseProblem enum constant %d>", p)
	return EnumData{GoName: goName, Name: name}
}

func (enum ParseProblem) GoString() string {
	return enum.Data().GoName
}

func (enum ParseProblem) String() string {
	return enum.Data().Name
}

func (enum ParseProblem) FormatString() string {
	return enum.Data().Format
}

var (
	_ fmt.GoStringer = ParseProblem(0)
	_ fmt.Stringer   = ParseProblem(0)
)
