package youyouayedee

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	type testRow struct {
		Input      string
		ExpectOK   bool
		ExpectUUID UUID
		ExpectErr  ErrParseFailed
	}

	sillyValue := UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	testData := [...]testRow{
		{
			Input:      "",
			ExpectOK:   true,
			ExpectUUID: NilUUID,
		},
		{
			Input:      "null",
			ExpectOK:   true,
			ExpectUUID: NilUUID,
		},
		{
			Input:      "nil",
			ExpectOK:   true,
			ExpectUUID: NilUUID,
		},
		{
			Input:      "max",
			ExpectOK:   true,
			ExpectUUID: MaxUUID,
		},
		{
			Input:      "00000000000010008000000000000000",
			ExpectOK:   true,
			ExpectUUID: sillyValue,
		},
		{
			Input:      "00000000-0000-1000-8000-000000000000",
			ExpectOK:   true,
			ExpectUUID: sillyValue,
		},
		{
			Input:      "{00000000-0000-1000-8000-000000000000}",
			ExpectOK:   true,
			ExpectUUID: sillyValue,
		},
		{
			Input:      "urn:uuid:00000000-0000-1000-8000-000000000000",
			ExpectOK:   true,
			ExpectUUID: sillyValue,
		},

		{
			Input:    "x",
			ExpectOK: false,
			ExpectErr: ErrParseFailed{
				Input:   []byte("x"),
				Problem: WrongLength,
				Args:    mkargs(uint(1)),
			},
		},
		{
			Input:    "0000000000001000800000000000000",
			ExpectOK: false,
			ExpectErr: ErrParseFailed{
				Input:   []byte("0000000000001000800000000000000"),
				Problem: WrongLength,
				Args:    mkargs(uint(31)),
			},
		},
		{
			Input:    "000000000000100080000000000000000",
			ExpectOK: false,
			ExpectErr: ErrParseFailed{
				Input:   []byte("000000000000100080000000000000000"),
				Problem: WrongLength,
				Args:    mkargs(uint(33)),
			},
		},
		{
			Input:    "0g000000000010008000000000000000",
			ExpectOK: false,
			ExpectErr: ErrParseFailed{
				Input:      []byte("0g000000000010008000000000000000"),
				Problem:    UnexpectedCharacter,
				Args:       mkargs(byte('g'), uint(1), "hex digit [0-9a-f]"),
				Index:      1,
				ActualByte: 'g',
			},
		},
		{
			Input:    "00000000:0000:1000:8000:000000000000",
			ExpectOK: false,
			ExpectErr: ErrParseFailed{
				Input:      []byte("00000000:0000:1000:8000:000000000000"),
				Problem:    UnexpectedCharacter,
				Args:       mkargs(byte(':'), uint(8), "'-'"),
				Index:      8,
				ExpectByte: '-',
				ActualByte: ':',
			},
		},
		{
			Input:    "00000000-0000-1000-c000-000000000000",
			ExpectOK: false,
			ExpectErr: ErrParseFailed{
				Input:      []byte("00000000-0000-1000-c000-000000000000"),
				Problem:    WrongVariant,
				Args:       mkargs(byte(0xc0), byte(0x80)),
				ExpectByte: 0x80,
				ActualByte: 0xc0,
			},
		},
	}

	for index, row := range testData {
		testName := fmt.Sprintf("%05d/%s", index, row.Input)
		t.Run(testName, func(t *testing.T) {
			uuid, err := Parse(row.Input)
			if row.ExpectOK {
				if err == nil && uuid == row.ExpectUUID {
					return
				}
				if err != nil {
					t.Errorf("Parse failed unexpectedly; error was: %v", err)
				} else {
					t.Errorf("Parse succeeded, but with unexpected value\n\texpect: %v\n\tactual: %v", row.ExpectUUID, uuid)
				}
			} else {
				xerr, ok := err.(ErrParseFailed)
				if ok && reflect.DeepEqual(xerr, row.ExpectErr) {
					return
				}
				if err == nil {
					t.Errorf("Parse succeeded unexpectedly; value was: %v", uuid)
				} else if !ok {
					t.Errorf("Parse failed as expected, but with error of unexpected type\n\texpect: %T\n\tactual: %T", row.ExpectErr, err)
				} else {
					t.Errorf("Parse failed as expected, but with unexpected error\n\texpect: %#v\n\tactual: %#v", row.ExpectErr, xerr)
				}
			}
		})
	}
}
