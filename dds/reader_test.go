package dds

import (
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	f, err := os.Open("de_dust2_radar.dds")
	if err != nil {
		t.Error(err)
	}

	c, err := DecodeConfig(f)
	if err != nil {
		t.Error(err)
	}
	t.Log(c)
}
