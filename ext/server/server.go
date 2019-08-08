package server

import (
	"context"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/client9/reopen"
	"github.com/kayac/go-katsubushi"
	"github.com/kohkimakimoto/cofu/support/logutil"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
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

func NewServer(config *Config) *Server {
	// setup logger
	logger := log.New("cofu-server")
	logger.SetLevel(log.INFO)
	logger.SetHeader(`${time_rfc3339} ${prefix} ${level}`)

	// set log level
	lv, err := logutil.LoglvlFromString(config.LogLevel)
	if err != nil {
		logger.Warn(err)
	}
	logger.SetLevel(lv)

	// create server instance
	srv := &Server{
		Config:  config,
		Logger:  logger,
		Echo:    echo.New(),
		DataDir: config.DataDir,
	}

	srv.Echo.Logger = logger
	srv.Echo.HideBanner = true
	srv.Echo.HidePort = true
	srv.Echo.Server.Addr = srv.Config.Addr

	return srv
}

func (srv *Server) Open() error {
	config := srv.Config
	logger := srv.Logger

	// open log
	if err := srv.openLogfile(); err != nil {
		return err
	}

	// Uniqid generator
	epoch, err := config.IDEpochTime()
	if err != nil {
		return err
	}
	katsubushi.Epoch = epoch
	gen, err := katsubushi.NewGenerator(0)
	if err != nil {
		return err
	}
	srv.Gen = gen

	// setup data directory as a temporary directory if it is not set.
	if srv.DataDir == "" {
		logger.Warn("Your 'data_dir' configuration is not set. Cofu server uses a temporary directory that is deleted after the process terminates.")

		tmpdir, err := ioutil.TempDir("", "cofu-server_data_")
		if err != nil {
			return err
		}
		logger.Warnf("Created temporary data directory: %s", tmpdir)
		srv.DataDir = tmpdir
		srv.UseTempDataDir = true
	}

	if _, err := os.Stat(srv.DataDir); os.IsNotExist(err) {
		err = os.MkdirAll(srv.DataDir, os.FileMode(0755))
		if err != nil {
			return err
		}
	}

	logger.Infof("Opened data directory: %s", srv.DataDir)

	// setup bolt database
	db, err := bolt.Open(srv.BoltDBPath(), 0600, nil)
	if err != nil {
		return err
	}
	srv.DB = db
	logger.Infof("Opened boltdb: %s", db.Path())

	return nil
}

func (srv *Server) openLogfile() error {
	if srv.Config.Logfile != "" {
		f, err := reopen.NewFileWriterMode(srv.Config.Logfile, 0644)
		if err != nil {
			return err
		}

		srv.Logger.SetOutput(f)
		srv.LogfileWriter = f
	} else {
		srv.LogfileWriter = reopen.Stdout
	}

	if srv.Config.AccessLogfile != "" {
		f, err := reopen.NewFileWriterMode(srv.Config.AccessLogfile, 0644)
		if err != nil {
			return err
		}
		srv.AccessLogFileWriter = f
	} else {
		srv.AccessLogFileWriter = reopen.Stdout
	}

	return nil
}

func (srv *Server) ListenAndServe() error {
	// open resources such as log files, database, temporary directory, etc.
	if err := srv.Open(); err != nil {
		return err
	}

	// Configure http servers (handlers and middleware)
	e := srv.Echo

	// error handler
	e.HTTPErrorHandler = errorHandler(srv)
	e.Use(ServerContextMiddleware(srv))
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Format:  `${time_rfc3339} ${remote_ip} ${host} ${method} ${uri} ${status} ${latency} ${latency_human} ${bytes_in} ${bytes_out}` + "\n",
		Output:  srv.AccessLogFileWriter,
	}))
	// handlers
	e.Any("/", IndexHandler)

	// handler for reopen logs
	go srv.sigusr1Handler()

	// start server.
	go func() {
		if err := e.Start(srv.Config.Addr); err != nil {
			e.Logger.Info(err)
		}
	}()

	srv.Logger.Infof("The server Listening on %s (pid: %d)", e.Server.Addr, os.Getpid())

	// see https://echo.labstack.com/cookbook/graceful-shutdown
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	srv.Logger.Infof("Received signal: %v", sig)
	timeout := time.Duration(srv.Config.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	e.Logger.Info("Shutting down the server")

	if err := e.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "fail to shutdown echo http server")
	}

	srv.Logger.Infof("Successfully shutdown")

	return nil
}

func (srv *Server) BoltDBPath() string {
	return filepath.Join(srv.DataDir, "server.bolt")
}

func (srv *Server) sigusr1Handler() {
	reopen := make(chan os.Signal, 1)
	signal.Notify(reopen, syscall.SIGUSR1)

	logger := srv.Logger

	for {
		select {
		case sig := <-reopen:
			logger.Infof("Received signal to reopen the logs: %v", sig)

			if err := srv.LogfileWriter.Reopen(); err != nil {
				logger.Error(fmt.Sprintf("failed to reopen log: %v", err))
			}

			if err := srv.AccessLogFileWriter.Reopen(); err != nil {
				logger.Error(fmt.Sprintf("failed to reopen access log: %v", err))
			}
		}
	}
}

func (srv *Server) Close() {
	if srv.DB != nil {
		if err := srv.DB.Close(); err != nil {
			srv.Logger.Error(err)
		}
	}

	if srv.UseTempDataDir {
		if err := os.RemoveAll(srv.DataDir); err != nil {
			srv.Logger.Error(err)
		}
		srv.Logger.Infof("Removed temporary directory: %s", srv.DataDir)
	}

	return
}

type ServerContext struct {
	echo.Context
	srv *Server
}

func (c *ServerContext) Server() *Server {
	return c.srv
}

func ServerContextMiddleware(srv *Server) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &ServerContext{c, srv}
			return next(cc)
		}
	}
}
