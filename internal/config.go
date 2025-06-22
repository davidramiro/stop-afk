package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	registryPath       = `SOFTWARE\WOW6432Node\Valve\cs2`
	registryKey        = "installpath"
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

var (
	cfgPath         = filepath.Join("game", "csgo", "cfg", "gamestate_integration_stop_afk.cfg")
	steamappsCSPath = filepath.Join("steamapps", "common", "Counter-Strike Global Offensive")
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
	cs2Path, err := c.findCS2()
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
	f, err := os.Open(filepath.Join(cs2Path, cfgPath))
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
	cs2Path, err := c.findCS2()
	if err != nil {
		return fmt.Errorf("could not check for cs2 directory: %w", err)
	}

	f, err := os.Create(filepath.Join(cs2Path, cfgPath))
	if err != nil {
		return fmt.Errorf("failed to create cfg file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, cfgContentTemplate, "1.0.0", 4242)
	if err != nil {
		return fmt.Errorf("failed to write cfg file contents: %w", err)
	}

	c.logCh <- LogMessage{LogSeverityOK, "gamestate config created, restart cs2 if running"}

	return nil
}
