package server

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"path/filepath"
	"time"
)

type Config struct {
	Server     *ServerConfig `toml:"server" json:"server"`
	configFile string        `toml:"-" json:"configFile"`
}

type ServerConfig struct {
	LogLevel      string `toml:"log_level" json:"log_level"`
	Addr          string `toml:"addr" json:"addr"`
	Logfile       string `toml:"log_file"`
	DataDir       string `toml:"data_dir"`
	AccessLogfile string `toml:"access_log_file"`
	ShutdownTimeout     int64  `toml:"shutdown_timeout"`
	IDEpoch       []int  `toml:"id_epoch" json:"id_epoch"`
}

const (
	DefaultPort = 5200
)

func NewConfig() *Config {
	return &Config{
		Server: &ServerConfig{
			LogLevel:      "info",
			Addr:          fmt.Sprintf("0.0.0.0:%d", DefaultPort),
			Logfile:       "",
			DataDir:       "",
			AccessLogfile: "",
			ShutdownTimeout: 10,
			IDEpoch:       []int{2019, 1, 1},
		},
		configFile: "",
	}
}

func (c *Config) LoadConfigFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	if err != nil {
		return err
	}

	if !filepath.IsAbs(c.Server.DataDir) {
		d, err := filepath.Abs(c.Server.DataDir)
		if err != nil {
			return err
		}
		c.Server.DataDir = d
	}

	c.configFile = path

	return nil
}

func (c *ServerConfig) IDEpochTime() (time.Time, error) {
	if len(c.IDEpoch) != 3 {
		return time.Now(), fmt.Errorf("id_epoch must be 3 int values")
	}

	return time.Date(c.IDEpoch[0], time.Month(c.IDEpoch[1]), c.IDEpoch[2], 0, 0, 0, 0, time.UTC), nil
}
