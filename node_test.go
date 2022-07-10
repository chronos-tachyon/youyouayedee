package youyouayedee

import (
	"fmt"
	"testing"
)

func TestNode(t *testing.T) {
	type testRow struct {
		Name    string
		Node    Node
		Zero    bool
		Global  bool
		Unicast bool
		GoStr   string
		Str     string
	}

	testData := [...]testRow{
		{
			Name:    "nil",
			Node:    NilNode,
			Zero:    true,
			Global:  false,
			Unicast: false,
			GoStr:   "youyouayedee.NilNode",
			Str:     "00:00:00:00:00:00",
		},
	}

	for index, row := range testData {
		testName := fmt.Sprintf("%05d/%s", index, row.Name)
		t.Run(testName, func(t *testing.T) {
			compare[bool](t, "IsZero", row.Zero, row.Node.IsZero())
			compare[bool](t, "IsGlobal", row.Global, row.Node.IsGlobal())
			compare[bool](t, "IsUnicast", row.Unicast, row.Node.IsUnicast())
			compare[string](t, "GoString", row.GoStr, row.Node.GoString())
			compare[string](t, "String", row.Str, row.Node.String())
		})
	}
}
