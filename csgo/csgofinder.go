package csgo

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
	"unicode"
)

func getCsgoPaths(libraryPaths []string) ([]string, error) {
	var csgoPaths []string
	for _, p := range libraryPaths {
		apppath := path.Join(p, "steamapps", "common")
		walkfun := func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			p = path.Clean(filepath.ToSlash(p))
			dir, file := path.Split(p)
			dir = path.Clean(dir)
			if info.IsDir() && p != apppath && dir != apppath {
				return filepath.SkipDir
			}
			if file == "steam_appid.txt" {
				b, err := ioutil.ReadFile(p)
				if err != nil {
					return err
				}

				b = bytes.TrimFunc(b, func(r rune) bool {
					return unicode.IsSpace(r) || r == rune(0)
				})
				if match, err := regexp.Match(`^730$`, b); match && err == nil {
					csgoPaths = append(csgoPaths, path.Join(filepath.ToSlash(p), ".."))
					return filepath.SkipDir
				}
			}
			return nil
		}
		err := filepath.Walk(apppath, walkfun)
		if err != nil {
			return nil, err
		}
	}

	return csgoPaths, nil
}

func getCsgoDemos(replayPaths []string, since time.Time) ([]string, error) {
	var demos []string
	for _, c := range replayPaths {
		paths, _ := filepath.Glob(path.Join(c, "*.dem"))
		for _, p := range paths {
			info, err := os.Stat(p)
			if err != nil {
				return nil, err
			}
			if info.ModTime().After(since) {
				demos = append(demos, path.Clean(filepath.ToSlash(p)))
			}
		}
	}
	return demos, nil
}
