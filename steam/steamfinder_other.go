// +build !windows

package steam

func getSteamPath() (string, error) {
    panic("getSteamPath not implemented for this platform")
}