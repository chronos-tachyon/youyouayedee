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

// MarshalBinary fulfills the "encoding".BinaryMarshaler interface.
func (uuid UUID) MarshalBinary() ([]byte, error) {
	return uuid[:], nil
}

// MarshalJSON fulfills the "encoding/json".Marshaler interface.
func (uuid UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(uuid.String())
}

// UnmarshalText fulfills the "encoding".TextUnmarshaler interface.
func (uuid *UUID) UnmarshalText(text []byte) error {
	var err error
	*uuid, err = parse(text, false)
	return err
}

// UnmarshalBinary fulfills the "encoding".BinaryUnmarshaler interface.
func (uuid *UUID) UnmarshalBinary(data []byte) error {
	var err error
	*uuid, err = parse(data, true)
	return err
}

// UnmarshalJSON fulfills the "encoding/json".Unmarshaler interface.
func (uuid *UUID) UnmarshalJSON(data []byte) error {
	*uuid = NilUUID

	if len(data) == 4 && string(data) == "null" {
		return nil
	}

	var str string
	err := json.Unmarshal(data, &str)
	if err == nil {
		var tmp [64]byte
		input := append(tmp[:0], str...)
		*uuid, err = parse(input, false)
	}
	return err
}

// Scan fulfills the "database/sql".Scanner interface.
func (uuid *UUID) Scan(value interface{}) error {
	var err error
	*uuid = NilUUID
	switch x := value.(type) {
	case nil:
		err = nil

	case string:
		var tmp [64]byte
		input := append(tmp[:0], x...)
		*uuid, err = parse(input, false)

	case []byte:
		*uuid, err = parse(x, true)

	default:
		err = fmt.Errorf("don't know how to interpret a value of type %T as a UUID", value)
	}
	return err
}

// Value fulfills the "database/sql/driver".Valuer interface.
func (uuid UUID) Value() (driver.Value, error) {
	return uuid[:], nil
}

// TextUUID is a wrapper type for UUID that SQL databases will store as a
// 36-character formatted string, instead of as a 16-byte raw binary value.
type TextUUID struct {
	UUID UUID
}

// Scan fulfills the "database/sql".Scanner interface.
func (text *TextUUID) Scan(value interface{}) error {
	return text.UUID.Scan(value)
}

// Value fulfills the "database/sql/driver".Valuer interface.
func (text TextUUID) Value() (driver.Value, error) {
	return text.UUID.String(), nil
}

var (
	_ encoding.TextMarshaler     = UUID{}
	_ encoding.BinaryMarshaler   = UUID{}
	_ json.Marshaler             = UUID{}
	_ driver.Valuer              = UUID{}
	_ driver.Valuer              = TextUUID{}
	_ encoding.TextUnmarshaler   = (*UUID)(nil)
	_ encoding.BinaryUnmarshaler = (*UUID)(nil)
	_ json.Unmarshaler           = (*UUID)(nil)
	_ sql.Scanner                = (*UUID)(nil)
	_ sql.Scanner                = (*TextUUID)(nil)
)
