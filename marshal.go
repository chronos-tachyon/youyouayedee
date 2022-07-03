package youyouayedee

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
)

// MarshalText fulfills the "encoding".TextMarshaler interface.
func (uuid UUID) MarshalText() ([]byte, error) {
	return uuid.AppendTo(make([]byte, 0, 36)), nil
}

// UnmarshalText fulfills the "encoding".TextUnmarshaler interface.
func (uuid *UUID) UnmarshalText(text []byte) error {
	var err error
	*uuid, err = ParseBytes(text)
	return err
}

// MarshalBinary fulfills the "encoding".BinaryMarshaler interface.
func (uuid UUID) MarshalBinary() ([]byte, error) {
	return uuid[:], nil
}

// UnmarshalBinary fulfills the "encoding".BinaryUnmarshaler interface.
func (uuid *UUID) UnmarshalBinary(data []byte) error {
	*uuid = NilUUID
	dataLen := uint(len(data))
	switch dataLen {
	case 0:
		return nil

	case 16:
		var tmp UUID
		copy(tmp[:], data)
		if !tmp.IsValid() {
			vb := tmp[8]
			return ParseError{
				Input:       string(data),
				Problem:     WrongVariant,
				Args:        []interface{}{vb},
				VariantByte: vb,
			}
		}
		*uuid = tmp
		return nil

	default:
		var err error
		*uuid, err = ParseBytes(data)
		if xerr, ok := err.(ParseError); ok && xerr.Problem == WrongLength {
			xerr.Problem = WrongBinaryLength
			err = xerr
		}
		return err
	}
}

// MarshalJSON fulfills the "encoding/json".Marshaler interface.
func (uuid UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.String())
}

// UnmarshalJSON fulfills the "encoding/json".Unmarshaler interface.
func (uuid *UUID) UnmarshalJSON(data []byte) error {
	if len(data) == 4 && string(data) == "null" {
		*uuid = NilUUID
		return nil
	}

	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		*uuid, err = Parse(str)
	}
	return err
}

// Value fulfills the "database/sql/driver".Valuer interface.
func (uuid UUID) Value() (driver.Value, error) {
	return uuid.MarshalBinary()
}

// Scan fulfills the "database/sql".Scanner interface.
func (uuid *UUID) Scan(value interface{}) error {
	var err error
	switch x := value.(type) {
	case nil:
		*uuid = NilUUID
		return nil

	case string:
		*uuid, err = Parse(x)
		return err

	case []byte:
		return uuid.UnmarshalBinary(x)

	default:
		return fmt.Errorf("don't know how to interpret a value of type %T as a UUID", value)
	}
}

// TextUUID is a wrapper type for UUID that SQL databases will store as a
// 36-character formatted string, instead of as a 16-byte raw binary value.
type TextUUID UUID

// Value fulfills the "database/sql/driver".Valuer interface.
func (uuid TextUUID) Value() (driver.Value, error) {
	return UUID(uuid).String(), nil
}

var (
	_ encoding.TextMarshaler     = UUID{}
	_ encoding.TextUnmarshaler   = (*UUID)(nil)
	_ encoding.BinaryMarshaler   = UUID{}
	_ encoding.BinaryUnmarshaler = (*UUID)(nil)
	_ json.Marshaler             = UUID{}
	_ json.Unmarshaler           = (*UUID)(nil)
	_ driver.Valuer              = UUID{}
	_ sql.Scanner                = (*UUID)(nil)
	_ driver.Valuer              = TextUUID{}
)
