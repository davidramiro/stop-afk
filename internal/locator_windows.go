//go:build windows

package internal

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
)

func (c *Config) findCS2() (string, error) {
	c.logCh <- LogMessage{LogSeverityInfo, "checking registry for cs2 install path"}
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("failed to open steam registry path: %w", err)
	}
	defer k.Close()

	p, _, err := k.GetStringValue(registryKey)
	if err != nil {
		return "", fmt.Errorf("failed to get cs2 install path: %w", err)
	}

	return p, nil
}
