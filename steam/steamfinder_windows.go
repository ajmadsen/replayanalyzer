package steam

import (
	"golang.org/x/sys/windows/registry"
	"path/filepath"
)

func GetInstallPath() (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}

	p, _, err := key.GetStringValue("SteamPath")
	if err != nil {
		return "", err
	}

	return filepath.Clean(filepath.ToSlash(p)), nil
}
