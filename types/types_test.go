package types

import "testing"

func TestRegex(t *testing.T) {
	c := Regex{
		Regex:      ".*",
		GroupRegex: "Foo",
	}
	err := c.Compile()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if c.TableName() != "regex" {
		t.Fatalf("Expected table name 'regex', got %s", c.TableName())
	}
	c.Kind = "collection"
	if c.TableName() != "collection_regex" {
		t.Fatalf("Expected table name 'regex', got %s", c.TableName())
	}
}
