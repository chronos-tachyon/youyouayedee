package youyouayedee

import (
	"fmt"
	"testing"
)

func TestConvert(t *testing.T) {
	type testRow struct {
		Name    string
		Input   UUID
		Version Version
		LSC     LeapSecondCalculator
		Output  UUID
		Err     error
	}

	testData := [...]testRow{
		{Name: "V1 to V1", Input: uuidV1, Version: 1, LSC: nil, Output: uuidV1},
		{Name: "V1 to V6", Input: uuidV1, Version: 6, LSC: nil, Output: uuidV6},
		{Name: "V1 to V7 nil LSC", Input: uuidV1, Version: 7, LSC: nil, Output: uuidV7A},
		{Name: "V1 to V7 dummy LSC", Input: uuidV1, Version: 7, LSC: lscDummy, Output: uuidV7A},
		{Name: "V1 to V7 fixed LSC", Input: uuidV1, Version: 7, LSC: lscFixed, Output: uuidV7B},
		{Name: "V1 to V8", Input: uuidV1, Version: 8, LSC: nil, Output: uuidV8FromV1},

		{Name: "V6 to V1", Input: uuidV6, Version: 1, LSC: nil, Output: uuidV1},
		{Name: "V6 to V6", Input: uuidV6, Version: 6, LSC: nil, Output: uuidV6},
		{Name: "V6 to V7 nil LSC", Input: uuidV6, Version: 7, LSC: nil, Output: uuidV7C},
		{Name: "V6 to V7 dummy LSC", Input: uuidV6, Version: 7, LSC: lscDummy, Output: uuidV7C},
		{Name: "V6 to V7 fixed LSC", Input: uuidV6, Version: 7, LSC: lscFixed, Output: uuidV7D},
		{Name: "V6 to V8", Input: uuidV6, Version: 8, LSC: nil, Output: uuidV8FromV6},

		{Name: "V7 to V1", Input: uuidV7A, Version: 1, LSC: nil, Err: ErrVersionMismatch{Requested: 1, Expected: []Version{7, 8}}},
		{Name: "V7 to V6", Input: uuidV7A, Version: 6, LSC: nil, Err: ErrVersionMismatch{Requested: 6, Expected: []Version{7, 8}}},
		{Name: "V7 to V7", Input: uuidV7A, Version: 7, LSC: nil, Output: uuidV7A},
		{Name: "V7 to V8", Input: uuidV7A, Version: 8, LSC: nil, Output: uuidV8FromV7A},

		{Name: "Nil to V3", Input: Nil, Version: 3, LSC: nil, Output: Nil},
		{Name: "Max to V3", Input: Max, Version: 3, LSC: nil, Output: Max},

		{Name: "Invalid to V3", Input: uuidInvalid, Version: 3, LSC: nil, Err: ErrInputNotValid{Input: uuidInvalid}},
		{Name: "V1 to V5", Input: uuidV1, Version: 5, LSC: nil, Err: ErrVersionMismatch{Requested: 5, Expected: []Version{1, 6, 7, 8}}},
		{Name: "V3 to V5", Input: uuidV3, Version: 5, LSC: nil, Err: ErrVersionMismatch{Requested: 5, Expected: []Version{3, 8}}},
		{Name: "V6 to V5", Input: uuidV6, Version: 5, LSC: nil, Err: ErrVersionMismatch{Requested: 5, Expected: []Version{1, 6, 7, 8}}},
	}

	for index, row := range testData {
		testName := fmt.Sprintf("%d/%s", index, row.Name)
		t.Run(testName, func(t *testing.T) {
			output, err := row.Input.Convert(row.Version, row.LSC)
			compare[UUID](t, "Convert", row.Output, output)
			compareError(t, "Convert", row.Err, err)
		})
	}
}
