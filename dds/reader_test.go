package dds

import (
	"image/png"
	"os"
	"testing"

	"path"
	"path/filepath"

	"image"

	"github.com/ajmadsen/replayanalyzer/csgo"
	"github.com/ajmadsen/replayanalyzer/steam"
)

func TestDecode(t *testing.T) {
	tstDir := "./test_output"
	err := os.RemoveAll(tstDir)
	if err != nil {
		t.Logf("failed to remove test directory")
	}
	for {
		// wait for folder to be deleted
		if _, err = os.Stat(tstDir); os.IsNotExist(err) {
			break
		}
	}
	err = os.Mkdir(tstDir, os.ModeDir)
	if err != nil {
		t.Fatal(err)
	}

	fnames := []string{
		"tests/smile_dxt1.dds",
		"tests/smile_dxt1a.dds",
		"tests/smile_dxt3.dds",
		"tests/smile_dxt5.dds",
		"tests/smile.rgba.dds",
	}

	for _, name := range fnames {
		f, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		c, err := DecodeConfig(f)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(c)

		f.Seek(0, os.SEEK_SET)

		i, err := Decode(f)
		if err != nil {
			t.Fatalf("could not decode directly: %v", err)
		}

		f.Seek(0, os.SEEK_SET)

		i, _, err = image.Decode(f)
		if err != nil {
			t.Fatalf("could not decode through image API: %v", err)
		}

		oname := path.Join(tstDir, filepath.Base(name)+".png")
		fo, err := os.Create(oname)
		if err != nil {
			t.Fatal(err)
		}
		defer fo.Close()
		err = png.Encode(fo, i)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestCSGOTextures(t *testing.T) {
	tstDir := "./test_output"
	err := os.RemoveAll(tstDir)
	if err != nil {
		t.Logf("failed to remove test directory")
	}
	for {
		// wait for folder to be deleted
		if _, err = os.Stat(tstDir); os.IsNotExist(err) {
			break
		}
	}
	err = os.Mkdir(tstDir, os.ModeDir)
	if err != nil {
		t.Fatal(err)
	}

	inst, err := steam.GetInstallPath()
	if err != nil {
		t.Skipf("not testing with csgo textures, could not find steam install: %v", err)
	}
	lPaths, err := steam.GetLibraryPaths(inst)
	if err != nil {
		t.Skipf("not testing with csgo textures, could not get library paths: %v", err)
	}
	csgos, err := csgo.GetInstallPaths(lPaths)
	if err != nil {
		t.Skipf("not testing with csgo textures, could not find csgo install: %v", err)
	}
	if len(csgos) == 0 {
		t.Skipf("not testing with csgo textures, could not find csgo install")
	}

	var overviews string
	for _, p := range csgos {
		p = path.Join(p, "csgo", "resource", "overviews")
		st, err := os.Stat(p)
		if err != nil {
			t.Logf("no folder found at %v", p)
			continue
		}
		if st.IsDir() {
			overviews = p
			break
		}
	}
	if overviews == "" {
		t.Skip("could not find an overviews folder to test")
	}

	t.Logf("found csgo overviews at %v", overviews)

	overFiles, err := filepath.Glob(path.Join(overviews, "*.dds"))
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range overFiles {
		t.Logf("converting texture %v", f)
		ff, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		defer ff.Close()

		i, err := Decode(ff)
		if err != nil {
			t.Fatalf("error decoding %v: %v", f, err)
		}

		oname := path.Join(tstDir, filepath.Base(f)+".png")
		fo, err := os.Create(oname)
		if err != nil {
			t.Fatal(err)
		}
		defer fo.Close()
		err = png.Encode(fo, i)
		if err != nil {
			t.Fatal(err)
		}
		fo.Close()
	}
}
