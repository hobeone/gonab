package main

import "testing"

func TestRegex(t *testing.T) {
	err := convertRegexes()
	if err != nil {
		t.Fatalf("Error %v", err)
	}

}
