package youyouayedee

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	type testRow struct {
		Name     string
		IsBinary bool
		Input    []byte
		Output   UUID
		Err      error
	}

	testData := [...]testRow{
		{
			Name:   "empty",
			Input:  strEmpty,
			Output: Nil,
		},
		{
			Name:   "null",
			Input:  strNull,
			Output: Nil,
		},
		{
			Name:   "nil",
			Input:  strNil,
			Output: Nil,
		},
		{
			Name:   "max",
			Input:  strMax,
			Output: Max,
		},
		{
			Name:     "binary",
			IsBinary: true,
			Input:    binV1AllZero,
			Output:   uuidV1AllZero,
		},
		{
			Name:   "text-no-dash",
			Input:  strV1AllZeroNoDash,
			Output: uuidV1AllZero,
		},
		{
			Name:   "text-with-dash",
			Input:  strV1AllZeroDash,
			Output: uuidV1AllZero,
		},
		{
			Name:   "text-with-dash-and-brace",
			Input:  strV1AllZeroBrace,
			Output: uuidV1AllZero,
		},
		{
			Name:   "text-with-dash-and-urn",
			Input:  strV1AllZeroURN,
			Output: uuidV1AllZero,
		},

		{
			Name:  "text-1x",
			Input: str1X,
			Err: ErrParseFailed{
				Input:   str1X,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(1)),
			},
		},
		{
			Name:  "text-3x",
			Input: str3X,
			Err: ErrParseFailed{
				Input:   str3X,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(3)),
			},
		},
		{
			Name:  "text-4x",
			Input: str4X,
			Err: ErrParseFailed{
				Input:   str4X,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(4)),
			},
		},
		{
			Name:  "text-no-dash-minus-one",
			Input: str31,
			Err: ErrParseFailed{
				Input:   str31,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(31)),
			},
		},
		{
			Name:  "text-no-dash-plus-one",
			Input: str33,
			Err: ErrParseFailed{
				Input:   str33,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(33)),
			},
		},
		{
			Name:  "text-with-dash-minus-one",
			Input: str35,
			Err: ErrParseFailed{
				Input:   str35,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(35)),
			},
		},
		{
			Name:  "text-with-dash-plus-one",
			Input: str37,
			Err: ErrParseFailed{
				Input:   str37,
				Problem: WrongBinaryLength,
				Args:    mkargs(uint(37)),
			},
		},

		{
			Name:  "text-with-dash-subst-g",
			Input: strG,
			Err: ErrParseFailed{
				Input:      strG,
				Problem:    UnexpectedCharacter,
				Args:       mkargs(byte('g'), uint(1), "hex digit [0-9a-f]"),
				Index:      1,
				ActualByte: 'g',
			},
		},
		{
			Name:  "text-with-colon",
			Input: strColon,
			Err: ErrParseFailed{
				Input:      strColon,
				Problem:    UnexpectedCharacter,
				Args:       mkargs(byte(':'), uint(8), "'-'"),
				Index:      8,
				ExpectByte: '-',
				ActualByte: ':',
			},
		},

		{
			Name:  "text-with-bad-variant-byte",
			Input: strVarC0,
			Err: ErrParseFailed{
				Input:      strVarC0,
				Problem:    WrongVariant,
				Args:       mkargs(byte(0xc0), byte(0x80)),
				ExpectByte: 0x80,
				ActualByte: 0xc0,
			},
		},
	}

	for index, row := range testData {
		testName := fmt.Sprintf("%05d/%s", index, row.Name)
		t.Run(testName, func(t *testing.T) {
			uuid, err := ParseBytes(row.Input)
			compare[UUID](t, "ParseBytes", row.Output, uuid)
			compareError(t, "ParseBytes", row.Err, err)

			if !row.IsBinary {
				expectErr := row.Err
				if xerr, ok := expectErr.(ErrParseFailed); ok {
					if xerr.Problem == WrongBinaryLength {
						xerr.Problem = WrongTextLength
						expectErr = xerr
					}
				}

				uuid, err = Parse(string(row.Input))
				compare[UUID](t, "Parse", row.Output, uuid)
				compareError(t, "Parse", expectErr, err)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	type testRow struct {
		Name    string
		UUID    UUID
		Zero    bool
		Max     bool
		Valid   bool
		Version Version
		Domain  DCEDomain
		ID      uint32
		GoStr   string
		Str     string
		URN     string
	}

	testData := [...]testRow{
		{
			Name:  "nil",
			UUID:  Nil,
			Zero:  true,
			Max:   false,
			Valid: false,
			GoStr: "youyouayedee.Nil",
			Str:   "00000000-0000-0000-0000-000000000000",
			URN:   "urn:uuid:00000000-0000-0000-0000-000000000000",
		},
		{
			Name:  "max",
			UUID:  Max,
			Zero:  false,
			Max:   true,
			Valid: false,
			GoStr: "youyouayedee.Max",
			Str:   "ffffffff-ffff-ffff-ffff-ffffffffffff",
			URN:   "urn:uuid:ffffffff-ffff-ffff-ffff-ffffffffffff",
		},
		{
			Name:    "V1 all zeroes",
			UUID:    uuidV1AllZero,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 1,
			GoStr:   "youyouayedee.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}",
			Str:     "00000000-0000-1000-8000-000000000000",
			URN:     "urn:uuid:00000000-0000-1000-8000-000000000000",
		},
		{
			Name:    "V1",
			UUID:    uuidV1,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 1,
			GoStr:   "youyouayedee.UUID{0xd3, 0xef, 0x76, 0x00, 0x6a, 0x95, 0x11, 0xec, 0x92, 0x34, 0x23, 0x58, 0x84, 0x0c, 0x40, 0xe6}",
			Str:     "d3ef7600-6a95-11ec-9234-2358840c40e6",
			URN:     "urn:uuid:d3ef7600-6a95-11ec-9234-2358840c40e6",
		},
		{
			Name:    "V2",
			UUID:    uuidV2,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 2,
			Domain:  Group,
			ID:      1000,
			GoStr:   "youyouayedee.UUID{0x00, 0x00, 0x03, 0xe8, 0x00, 0x00, 0x20, 0x00, 0x80, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}",
			Str:     "000003e8-0000-2000-8001-000000000000",
			URN:     "urn:uuid:000003e8-0000-2000-8001-000000000000",
		},
		{
			Name:    "V3",
			UUID:    uuidV3,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 3,
			GoStr:   "youyouayedee.UUID{0x8f, 0xbe, 0x4c, 0xf3, 0x9a, 0x53, 0x3a, 0x4f, 0xae, 0x1e, 0xa5, 0x87, 0x07, 0xfc, 0x4f, 0x4c}",
			Str:     "8fbe4cf3-9a53-3a4f-ae1e-a58707fc4f4c",
			URN:     "urn:uuid:8fbe4cf3-9a53-3a4f-ae1e-a58707fc4f4c",
		},
		{
			Name:    "V4",
			UUID:    uuidV4,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 4,
			GoStr:   "youyouayedee.UUID{0x90, 0x92, 0x58, 0x36, 0xe1, 0x23, 0x43, 0xf4, 0x8e, 0x5e, 0x64, 0x08, 0x6c, 0x46, 0xa7, 0xb6}",
			Str:     "90925836-e123-43f4-8e5e-64086c46a7b6",
			URN:     "urn:uuid:90925836-e123-43f4-8e5e-64086c46a7b6",
		},
		{
			Name:    "V5",
			UUID:    uuidV5,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 5,
			GoStr:   "youyouayedee.UUID{0xed, 0x81, 0xd1, 0xce, 0xb5, 0x5f, 0x58, 0xcc, 0xa2, 0xa4, 0xd5, 0xcf, 0x9b, 0x57, 0x88, 0x8f}",
			Str:     "ed81d1ce-b55f-58cc-a2a4-d5cf9b57888f",
			URN:     "urn:uuid:ed81d1ce-b55f-58cc-a2a4-d5cf9b57888f",
		},
		{
			Name:    "V6",
			UUID:    uuidV6,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 6,
			GoStr:   "youyouayedee.UUID{0x1e, 0xc6, 0xa9, 0x5d, 0x3e, 0xf7, 0x66, 0x00, 0x92, 0x34, 0x23, 0x58, 0x84, 0x0c, 0x40, 0xe6}",
			Str:     "1ec6a95d-3ef7-6600-9234-2358840c40e6",
			URN:     "urn:uuid:1ec6a95d-3ef7-6600-9234-2358840c40e6",
		},
		{
			Name:    "V7",
			UUID:    uuidV7B,
			Zero:    false,
			Max:     false,
			Valid:   true,
			Version: 7,
			GoStr:   "youyouayedee.UUID{0x01, 0x7e, 0x12, 0xef, 0x9c, 0x00, 0x70, 0x00, 0x80, 0x12, 0x34, 0x0a, 0x00, 0x1d, 0x33, 0x16}",
			Str:     "017e12ef-9c00-7000-8012-340a001d3316",
			URN:     "urn:uuid:017e12ef-9c00-7000-8012-340a001d3316",
		},
	}

	for index, row := range testData {
		testName := fmt.Sprintf("%05d/%s", index, row.Name)
		t.Run(testName, func(t *testing.T) {
			compare[bool](t, "IsZero", row.Zero, row.UUID.IsZero())
			compare[bool](t, "IsMax", row.Max, row.UUID.IsMax())
			compare[bool](t, "IsValid", row.Valid, row.UUID.IsValid())
			if row.Valid {
				compare[Version](t, "Version", row.Version, row.UUID.Version())
				if row.Version == 2 {
					compare[DCEDomain](t, "Domain", row.Domain, row.UUID.Domain())
					compare[uint32](t, "ID", row.ID, row.UUID.ID())
				}
			}
			compare[string](t, "GoString", row.GoStr, row.UUID.GoString())
			compare[string](t, "String", row.Str, row.UUID.String())
			compare[string](t, "URN", row.URN, row.UUID.URN())
		})
	}
}
