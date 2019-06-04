package agent

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"path/filepath"
	"time"
)

type Config struct {
	Agent      *AgentConfig   `toml:"agent" json:"agent"`
	Include    *IncludeConfig `toml:"include" json:"include"`
	configFile string         `toml:"-" json:"configFile"`
}

type AgentConfig struct {
	LogLevel           string   `toml:"log_level" json:"log_level"`
	Addr               string   `toml:"addr" json:"addr"`
	AuthorizedKeysFile string   `toml:"authorized_keys_file" json:"authorized_keys_file"`
	AuthorizedKeys     []string `toml:"authorized_keys" json:"authorized_keys"`
	DisableLocalAuth   bool     `toml:"disable_local_auth" json:"disable_local_auth"`
	HostKeyFile        string   `toml:"host_key_file" json:"-"`
	HostKey            string   `toml:"host_key" json:"-"`
	SandboxesDirectory string   `toml:"sandboxes_directory" json:"sandboxes_directory"`
	IDEpoch            []int    `toml:"id_epoch" json:"id_epoch"`
	HotReload          bool     `toml:"hot_reload" json:"hot_reload"`
}

type IncludeConfig struct {
	Files []string `toml:"files" json:"files"`
}

func NewConfig() *Config {
	return &Config{
		Agent: &AgentConfig{
			LogLevel:           "info",
			Addr:               fmt.Sprintf("0.0.0.0:%d", DefaultPort),
			AuthorizedKeysFile: "",
			AuthorizedKeys:     []string{},
			DisableLocalAuth:   false,
			HostKeyFile:        "",
			HostKey:            "",
			SandboxesDirectory: "/tmp/cofu-agent/sandboxes",
			IDEpoch:            []int{2019, 1, 1},
			HotReload:          false,
		},
		Include: &IncludeConfig{
			Files: []string{},
		},
		configFile: "",
	}
}

func (c *Config) LoadConfigFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	if err != nil {
		return err
	}

	for _, inc := range c.Include.Files {
		if err := c.includeConfigFile(inc); err != nil {
			return err
		}
	}

	if !filepath.IsAbs(c.Agent.SandboxesDirectory) {
		d, err := filepath.Abs(c.Agent.SandboxesDirectory)
		if err != nil {
			return err
		}
		c.Agent.SandboxesDirectory = d
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

const (
	DefaultPort = 2222
)

func (c *Config) includeConfigFile(include string) error {
	files, err := filepath.Glob(include)
	if err != nil {
		return err
	}

	for _, file := range files {
		_, err := toml.DecodeFile(file, c)
		if err != nil {
			return fmt.Errorf("failed loading included config file %s: %s", file, err)
		}
	}

	return nil
}

func (c *AgentConfig) IDEpochTime() (time.Time, error) {
	if len(c.IDEpoch) != 3 {
		return time.Now(), fmt.Errorf("id_epoch must be 3 int values")
	}

	return time.Date(c.IDEpoch[0], time.Month(c.IDEpoch[1]), c.IDEpoch[2], 0, 0, 0, 0, time.UTC), nil
}
