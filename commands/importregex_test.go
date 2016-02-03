package commands

import (
	"io/ioutil"
	"testing"
)

func TestNewsNabImport(t *testing.T) {
	contents, err := ioutil.ReadFile("nnregex")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	err = parseNewsNabRegex(contents)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}
