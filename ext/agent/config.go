package agent

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"path/filepath"
)

type Config struct {
	LogLevel           string   `toml:"log_level" json:"log_level"`
	Addr               string   `toml:"addr" json:"addr"`
	AuthorizedKeysFile string   `toml:"authorized_keys_file" json:"authorized_keys_file"`
	AuthorizedKeys     []string `toml:"authorized_keys" json:"authorized_keys"`
	DisableLocalAuth   bool     `toml:"disable_local_auth" json:"disable_local_auth"`
	HostKeyFile        string   `toml:"host_key_file" json:"-"`
	HostKey            string   `toml:"host_key" json:"-"`
	SandboxesDirectory string   `toml:"sandboxes_directory" json:"sandboxes_directory"`
	KeepSandboxes      int      `toml:"keep_sandboxes" json:"keep_sandboxes"`
	Environment        []string `toml:"environment" json:"environment"`
	EnvironmentFile    string   `toml:"environment_file" json:"environment_file"`
	HotReload          bool     `toml:"hot_reload" json:"hot_reload"`
	// Loaed config file
	configFile string `toml:"-" json:"configFile"`
}

const (
	DefaultAgentPort = 2222
)

func NewConfig() *Config {
	return &Config{
		LogLevel:           "info",
		Addr:               fmt.Sprintf("0.0.0.0:%d", DefaultAgentPort),
		AuthorizedKeysFile: "/etc/cofu-agent/authorized_keys",
		AuthorizedKeys:     []string{},
		DisableLocalAuth:   false,
		HostKeyFile:        "",
		HostKey:            "",
		SandboxesDirectory: "/tmp/cofu-agent/sandboxes",
		KeepSandboxes:      0,
		Environment:        []string{},
		EnvironmentFile:    "/etc/cofu-agent/environment",
		HotReload:          false,
		configFile:         "",
	}
}

func (c *Config) LoadConfigFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	if err != nil {
		return err
	}

	if !filepath.IsAbs(c.SandboxesDirectory) {
		d, err := filepath.Abs(c.SandboxesDirectory)
		if err != nil {
			return err
		}
		c.SandboxesDirectory = d
	}

	c.configFile = path

	return nil
}

func (c *Config) Reload() (*Config, error) {
	newConfig := NewConfig()
	if err := newConfig.LoadConfigFile(c.configFile); err != nil {
		return nil, err
	}

	return newConfig, nil
}
