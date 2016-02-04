package commands

import (
	"reflect"
	"testing"

	"github.com/hobeone/gonab/types"
)

func TestNewsNabToRegex(t *testing.T) {
	p := []string{
		"",
		"1",
		"misc.test",
		`/^(?P<name>.*?)\\s==\\s\\((?P<parts>\\d{1,3}\\/\\d{1,3})/i)`,
		"150",
		"1",
		"",
		"NULL",
	}
	reg, err := newzNabRegexToRegex(p)
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}
	expected := &types.Regex{
		ID:          1,
		Regex:       "/^(?P<name>.*?)\\\\s==\\\\s\\\\((?P<parts>\\\\d{1,3}\\\\/\\\\d{1,3})/i)",
		Description: "",
		Status:      true,
		Ordinal:     150,
		GroupName:   "misc.test",
	}
	if !reflect.DeepEqual(reg, expected) {
		t.Fatalf("Unexpected parse result %#v != %#v", reg, expected)
	}
}
