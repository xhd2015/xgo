package render

import (
	"bytes"
	"encoding/json"
	"testing"
)

// see https://github.com/xhd2015/xgo/issues/351
func TestReadStacksFromReader_LargeNumber(t *testing.T) {
	// Create a test case with a very large number in Results field
	testJSON := `{
		"Format": "stack",
		"Begin": "2024-03-01T12:00:00Z",
		"Children": [{
			"FuncInfo": {
				"Kind": "func",
				"Pkg": "test",
				"Name": "TestFunc"
			},
			"BeginNs": 1000,
			"EndNs": 2000,
			"Results": {
			    "BigInt":24804535128592856628882972797140548746494791012687994043373948793723059537920253648172292321648160710902664953408354783147738123454892060543908426629522664687519204924463335709306803386861532624573845766609055133697962124959618146314028128982094575718328507199178117054401501250326103136189258644305671834451667512335613221973678829157087399523755259396193907546838528903763882883751660577998454492461081951104423615227149584048652383674561490947441335090428607100533288626948641283887644492666480768848323929121609399830298075782015559286922042657941760394828548243453971784333332984038670960946556024954859483138003
			} 
		}]
	}`

	reader := bytes.NewReader([]byte(testJSON))
	stacks, _, err := readStacksFromReader(reader, true)
	if err != nil {
		t.Fatalf("readStacksFromReader failed: %v", err)
	}

	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}

	stack := stacks[0]
	if len(stack.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(stack.Children))
	}

	child := stack.Children[0]
	if child.FuncInfo == nil {
		t.Fatal("expected FuncInfo to be non-nil")
	}

	if child.FuncInfo.Name != "TestFunc" {
		t.Errorf("expected FuncInfo.Name to be 'TestFunc', got %q", child.FuncInfo.Name)
	}

	// The key test: verify that the large number was properly unmarshaled
	results, ok := child.Results.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Results to be map[string]interface{}, got %T", child.Results)
	}

	// this
	xgoRes, ok := results["BigInt"].(json.Number)
	if !ok {
		t.Fatalf("expected BigInt to be json.Number, got %T", results["BigInt"])
	}

	// Verify the number string matches exactly
	expected := "24804535128592856628882972797140548746494791012687994043373948793723059537920253648172292321648160710902664953408354783147738123454892060543908426629522664687519204924463335709306803386861532624573845766609055133697962124959618146314028128982094575718328507199178117054401501250326103136189258644305671834451667512335613221973678829157087399523755259396193907546838528903763882883751660577998454492461081951104423615227149584048652383674561490947441335090428607100533288626948641283887644492666480768848323929121609399830298075782015559286922042657941760394828548243453971784333332984038670960946556024954859483138003"
	if xgoRes != json.Number(expected) {
		t.Errorf("expected __xgo_res to be %q, got %q", expected, xgoRes)
	}
}
