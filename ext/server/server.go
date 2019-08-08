package server

import (
	"context"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/kohkimakimoto/cofu/support/logutil"
	"github.com/kohkimakimoto/hq/hq"
	"github.com/labstack/echo"
	"github.com/client9/reopen"
	"github.com/kayac/go-katsubushi"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(configFile string) error {
	config := NewConfig()

	if configFile == "" {
		return fmt.Errorf("-config-file is required option with -server.")
	}

	if err := config.LoadConfigFile(configFile); err != nil {
		return errors.Wrapf(err, "failed to load config from the file '%s'", configFile)
	}

	srv := NewServer(config)
	defer srv.Close()

	if err := srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

type Server struct {
	Config *Config
	// Logger
	Logger echo.Logger
	// LogfileWriter
	LogfileWriter reopen.Writer
	// LogLevel
	LogLevel log.Lvl
	// Echo web framework
	Echo *echo.Echo
	// AccessLog
	AccessLogFile *os.File
	// AccessLogFile
	AccessLogFileWriter reopen.Writer
	// DataDir
	DataDir string
	// UseTempDataDir
	UseTempDataDir bool
	// DB
	DB *bolt.DB
	// katsubushi
	Gen katsubushi.Generator

}

func NewServer(config ...*Config) *Server {
	var c *Config
	if len(config) == 0 {
		c = NewConfig()
	} else {
		c = config[0]
	}

	// create app instance
	srv := &Server{
		Config:  c,
		Echo:    echo.New(),
		DataDir: c.Server.DataDir,
	}

	srv.Echo.HideBanner = true
	srv.Echo.HidePort = true
	srv.Echo.Server.Addr = srv.Config.Server.Addr

	return srv
}

func (srv *Server) Open() error {
	config := srv.Config

	// log level
	lv, err := logutil.LoglvlFromString(config.Server.LogLevel)
	if err != nil {
		return err
	}
	srv.LogLevel = lv

	// setup logger
	logger := log.New(hq.Name)
	logger.SetLevel(srv.LogLevel)
	logger.SetHeader(`${time_rfc3339} ${level}`)
	srv.Logger = logger
	// setup echo logger
	srv.Echo.Logger = logger


	// Uniqid generator
	epoch, err := config.Server.IDEpochTime()
	if err != nil {
		return err
	}
	katsubushi.Epoch = epoch
	gen, err := katsubushi.NewGenerator(0)
	if err != nil {
		return err
	}
	srv.Gen = gen

	return nil
}

func (srv *Server) ListenAndServe() error {
	// open resources such as log files, database, temporary directory, etc.
	if err := srv.Open(); err != nil {
		return err
	}

	// Configure http servers (handlers and middleware)
	e := srv.Echo

	srv.Logger.Infof("The server Listening on %s (pid: %d)", e.Server.Addr, os.Getpid())

	// see https://echo.labstack.com/cookbook/graceful-shutdown
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	srv.Logger.Infof("Received signal: %v", sig)
	timeout := time.Duration(srv.Config.Server.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	e.Logger.Info("Shutting down the server")

	if err := e.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "fail to shutdown echo http server")
	}

	return nil
}

func (srv *Server) Close() {

}