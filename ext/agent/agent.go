package agent

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/support/logutil"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func Start(configFile string) error {
	config := NewConfig()

	if configFile == "" {
		return fmt.Errorf("-config-file is required option with -agent.")
	}

	if err := config.LoadConfigFile(configFile); err != nil {
		return errors.Wrapf(err, "failed to load config from the file '%s'", configFile)
	}

	a := NewAgent(config)
	defer a.Close()

	if err := a.Start(); err != nil {
		return err
	}

	return nil
}

type Agent struct {
	Config         *Config
	Logger         *log.Logger
	SessionManager *SessionManager
}

func NewAgent(config *Config) *Agent {
	// setup logger
	logger := log.New("cofu-agent")
	logger.SetLevel(log.INFO)
	logger.SetHeader(`${time_rfc3339} ${prefix} ${level}`)

	// set log level
	lv, err := logutil.LoglvlFromString(config.LogLevel)
	if err != nil {
		logger.Warn(err)
	}
	logger.SetLevel(lv)

	// create agent instance
	srv := &Agent{
		Config: config,
		Logger: logger,
	}

	return srv
}

func (a *Agent) Close() {
	// nothing to do.
}

func (a *Agent) Start() error {
	config := a.Config
	logger := a.Logger

	logger.Info("Starting cofu-agent")

	if config.SandboxesDirectory == "" {
		return fmt.Errorf("Require 'sandboxes_directory' config to sandoxes")
	}

	// Create Sandboxes Directory
	if _, err := os.Stat(config.SandboxesDirectory); os.IsNotExist(err) {
		err = os.MkdirAll(config.SandboxesDirectory, os.FileMode(0755))
		if err != nil {
			return err
		}
	}
	logger.Infof("Sandbox directory: %s", config.SandboxesDirectory)

	// session manager
	a.SessionManager = NewSessionManager(a)

	go func() {
		if err := startSSHServer(a); err != nil {
			logger.Info(err)
		}
	}()

	// wait signals
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down cofu-agent")

	return nil
}

func (a *Agent) CreateSandBoxDirIfNotExist(sess *Session) (string, error) {
	defaultUmask := syscall.Umask(0)
	defer syscall.Umask(defaultUmask)

	sandBoxDir := filepath.Join(a.Config.SandboxesDirectory, sess.SandboxName)
	if _, err := os.Stat(sandBoxDir); os.IsNotExist(err) {
		err = os.MkdirAll(sandBoxDir, os.FileMode(0777))
		if err != nil {
			return "", err
		}
	}

	return sandBoxDir, nil
}
