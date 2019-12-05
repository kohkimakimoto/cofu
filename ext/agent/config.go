package agent

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"path/filepath"
)

type Config struct {
	LogLevel           string                     `toml:"log_level" json:"log_level"`
	Addr               string                     `toml:"addr" json:"addr"`
	AuthorizedKeysFile string                     `toml:"authorized_keys_file" json:"authorized_keys_file"`
	AuthorizedKeys     []string                   `toml:"authorized_keys" json:"authorized_keys"`
	DisableLocalAuth   bool                       `toml:"disable_local_auth" json:"disable_local_auth"`
	HostKeyFile        string                     `toml:"host_key_file" json:"-"`
	HostKey            string                     `toml:"host_key" json:"-"`
	SandboxesDirectory string                     `toml:"sandboxes_directory" json:"sandboxes_directory"`
	KeepSandboxes      int                        `toml:"keep_sandboxes" json:"keep_sandboxes"`
	Environment        []string                   `toml:"environment" json:"environment"`
	EnvironmentFile    string                     `toml:"environment_file" json:"environment_file"`
	HotReload          bool                       `toml:"hot_reload" json:"hot_reload"`
	Functions          map[string]*FunctionConfig `toml:"functions" json:"services"`
	Include            *IncludeConfig             `toml:"include" json:"include"`
	// Loaed config file
	configFile string `toml:"-" json:"configFile"`
}

type FunctionConfig struct {
	// name
	Name string `toml:"-" json:"name"`
	// AuthorizedKeysFile
	AuthorizedKeysFile *string `toml:"authorized_keys_file" json:"authorized_keys_file"`
	// AuthorizedKeys
	AuthorizedKeys []string `toml:"authorized_keys" json:"authorized_keys"`
	// User
	User string `toml:"user" json:"user"`
	// Group
	Group string `toml:"group" json:"group"`
	// Entrypoint
	Entrypoint []string `toml:"entrypoint" json:"entrypoint"`
	// Command
	Command []string `toml:"command" json:"command"`
	// MaxProcesses
	MaxProcesses int `toml:"max_processes" json:"max_processes"`
	// Timeout
	Timeout int64 `toml:"timeout" json:"timeout"`
}

type IncludeConfig struct {
	Files []string `toml:"files" json:"files"`
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
		Functions:          map[string]*FunctionConfig{},
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

	for k, v := range c.Functions {
		v.Name = k
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

func (c *Config) Reload() (*Config, error) {
	newConfig := NewConfig()
	if err := newConfig.LoadConfigFile(c.configFile); err != nil {
		return nil, err
	}

	return newConfig, nil
}
