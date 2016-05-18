package csgo

import (
	"path/filepath"
	"testing"
)

func TestDemoReader(t *testing.T) {
	paths, _ := filepath.Glob("./tests/*.dem")
mainLoop:
	for _, p := range paths {
		r, err := NewReader(p)
		if err != nil {
			t.Error(err)
			continue
		}
		c := r.Subscribe("CSVSMsg_ServerInfo")
		err = r.Start()
		if err != nil {
			t.Error(err)
			continue
		}
		count := 0
		for msg := range c {
			count++
			if count != 1 {
				t.Error("more than one server info packet found")
				continue mainLoop
			}
			t.Log(msg.Body)
		}
	}
}
