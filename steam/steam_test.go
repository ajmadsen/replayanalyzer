package steam

import (
	"os"
	"path"
	"testing"
)

func TestGetSteamPath(t *testing.T) {
	steamPath, err := GetInstallPath()
	if err != nil {
		t.Error(err)
	}

	testPaths := []string{
		"config/config.vdf",
	}

	for _, p := range testPaths {
		p = path.Join(steamPath, p)
		_, err := os.Stat(p)
		if err != nil {
			t.Errorf("failed to stat %v: %v", p, err)
		}
	}
}

func TestGetLibraryPaths(t *testing.T) {
	steamPath, err := GetInstallPath()
	if err != nil {
		t.Error(err)
	}

	strings, err := GetLibraryPaths(steamPath)
	if err != nil {
		t.Error(err)
	}

	t.Logf("Found paths: %v\n", strings)

	for _, p := range strings {
		newPath := path.Join(p, "steamapps")
		info, err := os.Stat(newPath)
		if err != nil {
			t.Error(err)
		}
		if !info.IsDir() {
			t.Errorf("%v is not a directory despite being found by getLibraryPaths()", newPath)
		}
	}
}
