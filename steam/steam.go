package steam

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/andygrunwald/vdf"
)

var (
	keyMatcher = regexp.MustCompile(`^BaseInstallFolder_\d+`)
)

func GetLibraryPaths(steamPath string) ([]string, error) {
	var libraryPaths []string

	libraryPaths = append(libraryPaths, path.Clean(filepath.ToSlash(steamPath)))

	configFileStr := path.Join(steamPath, "config", "config.vdf")
	configFile, err := os.Open(configFileStr)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	configParser := vdf.NewParser(configFile)
	config, err := configParser.Parse()
	if err != nil {
		return nil, err
	}
	var ok bool

	cfgPath := []string{"InstallConfigStore", "Software", "Valve", "Steam"}
	nav := config
	for _, s := range cfgPath {
		if nav, ok = nav[s].(map[string]interface{}); !ok {
			return nil, fmt.Errorf("config.vdf missing %s", s)
		}
	}

	for k := range nav {
		if keyMatcher.Match([]byte(k)) {
			if p, ok := nav[k].(string); ok {
				libraryPaths = append(libraryPaths, path.Clean(filepath.ToSlash(p)))
			}
		}
	}

	return libraryPaths, nil
}
