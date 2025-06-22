//go:build linux

package internal

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
)

func (c *Config) findCS2() (string, error) {
	c.logCh <- LogMessage{LogSeverityInfo, "checking common linux steam installation paths"}

	u, err := user.Current()
	if err != nil {
		return "", err
	}
	home := u.HomeDir

	candidates := []string{
		filepath.Join(home, ".steam", "steam"),
		filepath.Join(home, ".steam", "root"),
		filepath.Join(home, ".local", "share", "Steam"),
		filepath.Join(home, ".var", "app", "com.valvesoftware.Steam", ".steam"),
	}

	for _, path := range candidates {
		realPath, err := filepath.EvalSymlinks(path)
		if err == nil {
			if _, err := os.Stat(filepath.Join(realPath, steamappsCSPath)); err == nil {
				return realPath + steamappsCSPath, nil
			}
		}
	}
	return "", errors.New("steam install folder not found")
}
