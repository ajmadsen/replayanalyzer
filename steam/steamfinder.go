// +build !windows

package steam

func GetInstallPath() (string, error) {
	panic("getSteamPath not implemented for this platform")
}
