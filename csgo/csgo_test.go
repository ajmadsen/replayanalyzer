package csgo

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var testTreePaths = []string{
	"Dsteamapps",
	"Dsteamapps/common",
	"Dsteamapps/common/game1",
	"Fsteamapps/common/game1/steam_appid.txt\n730",
	"Dsteamapps/common/game1/csgo",
	"Dsteamapps/common/game1/csgo/replays",
	"Fsteamapps/common/game1/csgo/replays/12345.dem",
	"Fsteamapps/common/game1/csgo/replays/34567.dem",
	"Dsteamapps/common/game2",
	"Fsteamapps/common/game2/steam_appid.txt\n7300",
	"Dsteamapps/common/game3",
	"Fsteamapps/common/game3/steam_appid.txt\n1730",
	"Dsteamapps/common/game4",
	"Fsteamapps/common/game4/steam_appid.txt",
	"Dsteamapps/common/game4/csgo",
	"Dsteamapps/common/game4/csgo/replays",
	"Fsteamapps/common/game4/csgo/replays/23456.dem",
	"Dsteamapps/common/game5",
	"Dsteamapps/common/game5/subdir",
	"Fsteamapps/common/game5/subdir/steam_appid.txt\n730",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomString() string {
	charSet := "abcdefghijklmnopqrstuvwxyz0123456789"
	s := ""
	for i := 0; i < 8; i++ {
		s += string(charSet[rand.Intn(len(charSet))])
	}
	return s
}

func makeTestTree(paths []string) (string, error) {
	tmpDir := filepath.ToSlash(path.Join(os.TempDir(), getRandomString()))
	err := os.MkdirAll(tmpDir, os.ModeDir)
	if err != nil {
		return "", err
	}

	for _, p := range paths {
		t := p[0]
		p = p[1:]
		switch t {
		case 'D':
			dirName := path.Join(tmpDir, p)
			err = os.Mkdir(dirName, os.ModeDir)
			if err != nil {
				return "", err
			}
		case 'F':
			nl := strings.Index(p, "\n")
			var fileName, contents string
			if nl > 0 {
				fileName = p[:nl]
				contents = p[nl+1:]
			} else {
				fileName = p
			}
			fileName = path.Join(tmpDir, fileName)
			err = ioutil.WriteFile(fileName, []byte(contents), 0644)
			if err != nil {
				return "", err
			}
		default:
			return "", fmt.Errorf("invalid type %v", t)
		}
	}

	return tmpDir, nil
}

func TestGetCsgoPaths(t *testing.T) {
	var testPaths []string
	solutions := map[string]bool{}
	for i := 0; i < 5; i++ {
		tp, err := makeTestTree(testTreePaths)
		if err != nil {
			t.Error(err)
		}
		defer os.RemoveAll(tp)
		testPaths = append(testPaths, tp)
		solutions[path.Join(tp, "steamapps/common/game1")] = true
	}

	ps, err := GetInstallPaths(testPaths)
	if err != nil {
		t.Error(err)
	}

	revmap := map[string]bool{}
	for _, p := range ps {
		if !solutions[p] {
			t.Errorf("returned %v which does not look like CSGO", p)
		}
		revmap[p] = true
	}
	for s := range solutions {
		if !revmap[s] {
			t.Errorf("did not return path %v", s)
		}
	}
}

func TestGetCsgoDemos(t *testing.T) {
	var testPaths []string
	solutions := map[string]bool{}
	for i := 0; i < 5; i++ {
		tp, err := makeTestTree(testTreePaths)
		if err != nil {
			t.Error(err)
		}
		defer os.RemoveAll(tp)
		tp = path.Join(tp, "steamapps", "common", "game1", "csgo", "replays")
		testPaths = append(testPaths, tp)
		for _, p := range []string{"12345.dem", "34567.dem"} {
			p = path.Join(tp, p)
			solutions[p] = true
		}
	}

	demos, err := GetDemos(testPaths, time.Time{})
	if err != nil {
		t.Error(err)
	}

	revmap := map[string]bool{}
	for _, d := range demos {
		if !solutions[d] {
			t.Errorf("found a demo file outside of tree %v", d)
		}
		revmap[d] = true
	}
	for s := range solutions {
		if !revmap[s] {
			t.Errorf("did not find demo %v", s)
		}
	}
}
