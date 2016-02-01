package config

import "testing"

func TestReadSampleConfig(t *testing.T) {
	c := NewConfig()
	err := c.ReadConfig("../config_sample.json")
	if err != nil {
		t.Fatalf("Error reading sample config: %v", err)
	}
}
