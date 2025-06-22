package internal

import (
	"errors"
	"fmt"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

const (
	registryPath       = `SOFTWARE\WOW6432Node\Valve\cs2`
	registryKey        = "installpath"
	steamappsCSPath    = "steamapps/common/Counter-Strike Global Offensive"
	cfgPath            = `\game\csgo\cfg\gamestate_integration_stop_afk.cfg`
	cfgContentTemplate = `"Stop AFK v%s"
{
 "uri" "http://127.0.0.1:%d"
 "timeout" "5.0"
 "buffer" "0.1"
 "throttle" "0.5"
 "heartbeat" "0"
 "data"
 {
   "round"          "1" // round phase, bomb state and round winner
 }
}
`
)

type Config struct {
	logCh chan<- LogMessage
}

func NewConfig(logCh chan<- LogMessage) *Config {
	return &Config{
		logCh: logCh,
	}
}

func (c *Config) Init(version string, port int) {
	cs2Path, err := c.getInstallPath()
	if err != nil {
		c.logCh <- LogMessage{LogSeverityFail, "cs2 installation not found."}
	}

	c.logCh <- LogMessage{LogSeverityOK, "cs2 found at " + cs2Path}

	i, err := c.checkConfig(cs2Path, version, port)
	if err != nil {
		c.logCh <- LogMessage{LogSeverityFail, "failed to check config: " + err.Error()}
	}

	if !i {
		err = c.createConfig()
		if err != nil {
			c.logCh <- LogMessage{LogSeverityFail, "failed to create config: " + err.Error()}
		}
	}

	c.logCh <- LogMessage{LogSeverityOK, "finished config setup"}

}

func (c *Config) checkConfig(cs2Path string, version string, port int) (bool, error) {
	f, err := os.Open(cs2Path + cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.logCh <- LogMessage{LogSeverityOK, "gamestate config not found, creating one"}
			return false, nil
		}
		return false, fmt.Errorf("failed to check for cfg file: %w", err)
	}

	defer f.Close()

	c.logCh <- LogMessage{LogSeverityOK, "gamestate config found, checking if it is up to date"}
	content := make([]byte, 20)
	_, err = f.Read(content)
	if err != nil {
		return false, fmt.Errorf("failed to read cfg file: %w", err)
	}

	if string(content) != fmt.Sprintf(cfgContentTemplate, version, port)[:20] {
		c.logCh <- LogMessage{LogSeverityInfo, "gamestate config outdated, updating"}
		return false, nil
	}

	return true, nil
}

func (c *Config) createConfig() error {
	cs2Path, err := c.getInstallPath()
	if err != nil {
		return fmt.Errorf("could not check for cs2 directory: %w", err)
	}

	f, err := os.Create(cs2Path + cfgPath)
	if err != nil {
		return fmt.Errorf("failed to create cfg file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, cfgContentTemplate, "1.0.0", 4242)
	if err != nil {
		return fmt.Errorf("failed to write cfg file contents: %w", err)
	}

	return nil
}

func (c *Config) getInstallPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return c.findCS2Windows()
	case "linux":
		return c.findCS2Linux()
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)

	}
}

func (c *Config) findCS2Windows() (string, error) {
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

func (c *Config) findCS2Linux() (string, error) {
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
