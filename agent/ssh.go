package agent

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func startSSHServer(a *Agent) error {
	logger := a.Logger

	// ssh handlers
	ssh.Handle(func(sess ssh.Session) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
				io.WriteString(sess.Stderr(), fmt.Sprintf("%v\r\n", err))
				sess.Exit(1)
			}
		}()

		logger.Infof("Connected from: %s", sess.RemoteAddr().String())

		err := handleSSHSession(a, sess)
		if err != nil {
			io.WriteString(sess.Stderr(), fmt.Sprintf("ERROR: %v\r\n", err))
			logger.Errorf("%v", err)
		}

		status := exitStatus(err)
		sess.Exit(status)
		logger.Infof("SSH connection exited with status %d", status)
	})

	var options []ssh.Option

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		if a.Config.Agent.HotReload {
			config, err := a.Config.Reload()
			if err != nil {
				logger.Errorf("failed to reload config: %v", err)
			}
			a.Config = config
		}

		remoteAddr := ctx.RemoteAddr().String()
		remoteHost := remoteAddr[:strings.LastIndex(remoteAddr, ":")]

		if a.Config.Agent.DisableLocalAuth && (remoteHost == "127.0.0.1" || remoteHost == "[::1]") {
			logger.Debugf("passed public key auth because the config disables local auth")
			return true
		}

		if checkAuthKey(a, ctx, key) {
			return true
		}

		return false
	})

	options = append(options, publicKeyOption)

	if a.Config.Agent.HostKeyFile != "" {
		logger.Infof("Using host key file %s", a.Config.Agent.HostKeyFile)

		if _, err := os.Stat(a.Config.Agent.HostKeyFile); os.IsNotExist(err) {
			b, err := generateNewKey()
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(a.Config.Agent.HostKeyFile, b, 0600); err != nil {
				return err
			}
			logger.Infof("%s is not existed. generated it automatically.", a.Config.Agent.HostKeyFile)
		}

		hostKeyOption := ssh.HostKeyFile(a.Config.Agent.HostKeyFile)
		options = append(options, hostKeyOption)

		logger.Infof("Using host key file %s", a.Config.Agent.HostKeyFile)
	} else if a.Config.Agent.HostKey != "" {
		hostKeyOption := ssh.HostKeyPEM([]byte(a.Config.Agent.HostKey))
		options = append(options, hostKeyOption)

		logger.Info("Using host key from the config file")
	} else {
		// generated host key for development environment.
		hostKeyFile := "/tmp/cofu-agent/host.key"
		if _, err := os.Stat(hostKeyFile); os.IsNotExist(err) {
			b, err := generateNewKey()
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(hostKeyFile, b, 0600); err != nil {
				return err
			}
		}
		hostKeyOption := ssh.HostKeyFile(hostKeyFile)
		options = append(options, hostKeyOption)

		logger.Infof("Using host key file %s", hostKeyFile)
	}

	logger.Infof("Starting SSH protocol server on %s", a.Config.Agent.Addr)

	// start ssh server
	return ssh.ListenAndServe(a.Config.Agent.Addr, nil, options...)
}

func handleSSHSession(a *Agent, sshSession ssh.Session) error {

	return nil
}

func handleStatus(a *Agent, sess *Session) error {
	resp := map[string]interface{}{
		"config": a.Config,
	}

	return sess.JSONPretty(resp)
}

func exitStatus(err error) int {
	var exitStatus int
	if err != nil {
		if e2, ok := err.(*exec.ExitError); ok {
			if status, ok := e2.Sys().(syscall.WaitStatus); ok {
				exitStatus = status.ExitStatus()
			} else {
				exitStatus = 1
			}
		} else {
			exitStatus = 1
		}
	} else {
		exitStatus = 0
	}

	return exitStatus
}

func checkAuthKey(a *Agent, ctx ssh.Context, key ssh.PublicKey) bool {
	config := a.Config.Agent
	logger := a.Logger

	var keysdata []byte

	authorizedKeysFile := config.AuthorizedKeysFile

	if authorizedKeysFile != "" {
		data, err := ioutil.ReadFile(authorizedKeysFile)
		if err != nil {
			logger.Error(err)
			return false
		}

		keysdata = data
	}

	authorizedKeys := config.AuthorizedKeys

	for _, s := range authorizedKeys {
		keysdata = append(keysdata, []byte(s+"\n")...)
	}

	scanner := bufio.NewScanner(bytes.NewBuffer(keysdata))
	for scanner.Scan() {
		keyLine := scanner.Bytes()
		if len(keyLine) == 0 {
			continue
		}

		allowed, _, _, _, err := ssh.ParseAuthorizedKey(keyLine)
		if err != nil {
			logger.Error(err)
			return false
		}

		ok := ssh.KeysEqual(key, allowed)
		if ok {
			logger.Debugf("authed key: %s", string(keyLine))
			return true
		}
	}

	return false
}

func generateNewKey() ([]byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	buf := &bytes.Buffer{}
	if err := pem.Encode(buf, privateKey); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
