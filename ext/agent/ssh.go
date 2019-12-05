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
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/cofu/ext/agent/envfile"
	"github.com/kr/pty"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

func startSSHServer(a *Agent) error {
	logger := a.Logger

	// ssh handlers
	ssh.Handle(func(sess ssh.Session) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error(err)
				if _, err := io.WriteString(sess.Stderr(), fmt.Sprintf("%v\r\n", err)); err != nil {
					logger.Error(err)
				}

				if err := sess.Exit(1); err != nil {
					logger.Error(err)
				}
			}
		}()

		logger.Infof("Connected from: %s", sess.RemoteAddr().String())

		err := handleSSHSession(a, sess)
		if err != nil {
			if _, err := io.WriteString(sess.Stderr(), fmt.Sprintf("ERROR: %v\r\n", err)); err != nil {
				logger.Error(err)
			}
			logger.Error(err)
		}

		status := exitStatus(err)
		if err := sess.Exit(status); err != nil {
			logger.Error(err)
		}
		logger.Infof("SSH connection exited with status %d", status)
	})

	var options []ssh.Option

	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		if a.Config.HotReload {
			config, err := a.Config.Reload()
			if err != nil {
				logger.Errorf("failed to reload config: %v", err)
			}
			a.Config = config
		}

		remoteAddr := ctx.RemoteAddr().String()
		remoteHost := remoteAddr[:strings.LastIndex(remoteAddr, ":")]

		logger.Debugf("remoteAddr: %s", remoteAddr)
		logger.Debugf("remoteHost: %s", remoteHost)

		if a.Config.DisableLocalAuth && (remoteHost == "127.0.0.1" || remoteHost == "[::1]") {
			logger.Debugf("passed public key auth because the config disables local auth")
			return true
		}

		if checkAuthKey(a, ctx, key) {
			return true
		}

		return false
	})

	options = append(options, publicKeyOption)

	if a.Config.HostKeyFile != "" {
		if _, err := os.Stat(a.Config.HostKeyFile); os.IsNotExist(err) {
			b, err := generateNewKey()
			if err != nil {
				return err
			}
			if err := ioutil.WriteFile(a.Config.HostKeyFile, b, 0600); err != nil {
				return err
			}
			logger.Infof("%s is not existed. generated it automatically.", a.Config.HostKeyFile)
		}

		hostKeyOption := ssh.HostKeyFile(a.Config.HostKeyFile)
		options = append(options, hostKeyOption)

		logger.Infof("Using host key file %s", a.Config.HostKeyFile)
	} else if a.Config.HostKey != "" {
		hostKeyOption := ssh.HostKeyPEM([]byte(a.Config.HostKey))
		options = append(options, hostKeyOption)

		logger.Info("Using host key from the config file")
	} else {
		// generated host key for development environment.
		hostKeyFile := "/tmp/cofu-agent-host.key"
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

	logger.Infof("Starting SSH protocol server on %s", a.Config.Addr)

	// start ssh server
	return ssh.ListenAndServe(a.Config.Addr, nil, options...)
}

func handleSSHSession(a *Agent, sshSession ssh.Session) error {
	logger := a.Logger

	sess := NewSession(a, sshSession)
	if err := a.SessionManager.SetSession(sess); err != nil {
		return err
	}
	defer sess.Terminate()

	logger.Infof("Allocated session: %s sandbox: %s", sess.ID, sess.Sandbox)

	ptyReq, winCh, isPty := sess.Pty()

	logger.Debugf("cofu-agent got a ssh command: %v", sess.Command())
	logger.Debugf("The session user is %s", sess.User())

	currentEnviron := append(sess.Environ(),
		fmt.Sprintf("COFU_AGENT_VERSION=%s", cofu.Version),
		fmt.Sprintf("COFU_AGENT_SESSION=%s", sess.ID),
		fmt.Sprintf("COFU_AGENT_SESSION_ID=%s", sess.ID),
		fmt.Sprintf("COFU_AGENT_SANDBOX=%s", sess.Sandbox),
		fmt.Sprintf("COFU_COMMAND=%s", cofu.BinPath),
	)

	if sess.Fn != nil {
		logger.Infof("fetched function: %s", sess.Fn.Name)
		currentEnviron = append(currentEnviron, fmt.Sprintf("COFU_AGENT_FUNCTION=%s", sess.Fn.Name))
	}

	sessCommand := []string{}
	if len(sess.Command()) > 0 {
		sessCommand = sess.Command()
	} else {
		if sess.Fn != nil {
			if sess.Fn.Command != nil && len(sess.Fn.Command) > 0 {
				sessCommand = sess.Fn.Command
			}
		} else {
			sessCommand = []string{"/bin/bash"}
		}
	}

	if sess.Fn != nil {
		currentEnviron = append(currentEnviron, fmt.Sprintf("COFU_AGENT_FUNCTION_SESSION_COMMAND=%s", strings.Join(sessCommand, " ")))
	}

	if isPty {
		currentEnviron = append(currentEnviron, "COFU_AGENT_PTY=1")
	}

	// setup user and group
	uid, gid, u, err := getUidAndGid(sess)
	if err != nil {
		return err
	}
	logger.Debugf("This session runs by user: %d, group: %d", uid, gid)

	if os.Getuid() != 0 && os.Getuid() != uid {
		return errors.New("The Cofu Agent is running non root user. The connection can be accepted only by same user who runs Cofu Agent.")
	}

	sess.Uid = uid
	sess.Gid = gid
	sess.UserStruct = u

	// set environment about user
	// This logic was borrowed from `do_setup_env` of session.c in the openssh.
	currentEnviron = append(currentEnviron,
		fmt.Sprintf("USER=%s", u.Name),
		fmt.Sprintf("LOGNAME=%s", u.Name),
		fmt.Sprintf("HOME=%s", u.HomeDir),
	)

	// support forward agent
	currentEnviron, err = takeForwardAgentIfRequested(sess, currentEnviron)
	if err != nil {
		return err
	}

	// setup sandbox
	sandBoxDir, err := a.CreateSandBoxDirIfNotExist(sess)
	if err != nil {
		return err
	}
	currentEnviron = append(currentEnviron, fmt.Sprintf("COFU_AGENT_SANDBOX_DIR=%s", sandBoxDir))

	// setup environment variables
	if a.Config.EnvironmentFile != "" {
		if _, err := os.Stat(a.Config.EnvironmentFile); err == nil {
			appendedEnv, err := envfile.Load(a.Config.EnvironmentFile, currentEnviron)
			if err != nil {
				logger.Error(err)
			}

			if appendedEnv != nil {
				currentEnviron = append(currentEnviron, appendedEnv...)
			}
		}
	}

	for _, v := range a.Config.Environment {
		currentEnviron = append(currentEnviron, expandEnvironToString(v, currentEnviron))
	}

	commandAndArgs := []string{}
	if sess.Fn != nil && sess.Fn.Entrypoint != nil && len(sess.Fn.Entrypoint) > 0 {
		commandAndArgs = sess.Fn.Entrypoint
	} else {
		commandAndArgs = []string{}
	}
	commandAndArgs = append(commandAndArgs, sessCommand...)
	if len(commandAndArgs) == 0 {
		return errors.New("You've successfully authenticated, but it has no command to run.")
	}

	command := commandAndArgs[0]
	args := []string{}
	if len(commandAndArgs) >= 2 {
		args = commandAndArgs[1:]
	}

	logger.Debugf("command: %v", commandAndArgs)
	// exec command
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), currentEnviron...)
	if os.Getuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}
	cmd.Dir = sandBoxDir

	if isPty {
		cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
		f, err := pty.Start(cmd)
		if err != nil {
			return err
		}

		go func() {
			for win := range winCh {
				setWinsize(f, win.Width, win.Height)
			}
		}()

		go func() {
			io.Copy(f, sess) // stdin
		}()

		io.Copy(sess, f) // stdout

		return cmd.Wait()
	} else {
		cmd.Stdout = sess
		cmd.Stderr = sess.Stderr()

		// get stdin as a pipe
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			io.Copy(stdin, sess)
			stdin.Close()
		}()

		if err := cmd.Start(); err != nil {
			return err
		}

		if sess.Fn != nil && sess.Fn.Timeout > 0 {
			done := make(chan error)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-time.After(time.Duration(sess.Fn.Timeout) * time.Second):
				if err := cmd.Process.Kill(); err != nil {
					return errors.Wrapf(err, "failed to kill: "+err.Error())
				}
				return fmt.Errorf("Cofu Agent function timeout. it took time over %d sec.", sess.Fn.Timeout)
			case err := <-done:
				defer close(done)
				if err != nil {
					return err
				}
				return nil
			}
		} else {
			return cmd.Wait()
		}
	}
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

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func getUidAndGid(sess *Session) (int, int, *user.User, error) {
	if sess.Fn != nil {
		return getUidAndGidFromFunctionConfig(sess.Fn)
	} else {
		return getUidAndGidFromUsername(sess.User())
	}
}

func getUidAndGidFromFunctionConfig(fn *FunctionConfig) (int, int, *user.User, error) {
	var uid, gid int
	var u *user.User
	var err error
	if fn.User != "" {
		uid, gid, u, err = getUidAndGidFromUsername(fn.User)
		if err != nil {
			return -1, -1, nil, err
		}
	} else {
		uid = os.Getuid()
		gid = os.Getgid()

		u, err = LookupUserStruct(strconv.Itoa(uid))
		if err != nil {
			return -1, -1, nil, err
		}
	}

	groupName := fn.Group
	if groupName != "" {
		id, err := LookupGroup(groupName)
		if err != nil {
			return -1, -1, nil, err
		}
		gid = id
	}

	return uid, gid, u, nil
}

func getUidAndGidFromUsername(userName string) (int, int, *user.User, error) {
	u, err := LookupUserStruct(userName)
	if err != nil {
		return -1, -1, nil, err
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, -1, nil, err
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return -1, -1, nil, err
	}

	return uid, gid, u, nil
}

func takeForwardAgentIfRequested(sess *Session, env []string) ([]string, error) {
	logger := sess.Agt.Logger

	if ssh.AgentRequested(sess) {
		l, tmpDir, err := NewAgentListener(sess)
		if err != nil {
			return env, err
		}
		logger.Debugf("Created agent listener to forward agent: %s", l.Addr().String())

		sess.ForwardAgentListener = l
		go func() {
			ssh.ForwardAgentConnections(l, sess)
			logger.Debugf("Closed agent listener: %s", l.Addr().String())
			os.RemoveAll(tmpDir)
			logger.Debugf("Removed: %s", tmpDir)
		}()

		env = append(env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
		return env, nil
	}

	return env, nil
}

// I borrowed this code from https://github.com/gliderlabs/ssh/blob/47df570d18ad49f77cf66f76bc3fce3e92400768/agent.go#L38
// And modify code to remove temporary directory after closing
func NewAgentListener(sess *Session) (net.Listener, string, error) {
	dir, err := ioutil.TempDir("", "auth-agent")
	if err != nil {
		return nil, dir, err
	}

	if err := os.Chown(dir, sess.Uid, sess.Gid); err != nil {
		return nil, dir, err
	}

	l, err := net.Listen("unix", path.Join(dir, "listener.sock"))
	if err != nil {
		os.RemoveAll(dir)
		return nil, dir, err
	}

	if err := os.Chown(l.Addr().String(), sess.Uid, sess.Gid); err != nil {
		return nil, dir, err
	}

	return l, dir, nil
}

func LookupGroup(id string) (int, error) {
	var g *user.Group

	if _, err := strconv.Atoi(id); err == nil {
		g, err = user.LookupGroupId(id)
		if err != nil {
			return -1, err
		}
	} else {
		g, err = user.LookupGroup(id)
		if err != nil {
			return -1, err
		}
	}

	return strconv.Atoi(g.Gid)
}

func LookupUser(id string) (int, error) {
	u, err := LookupUserStruct(id)
	if err != nil {
		return -1, err
	}

	return strconv.Atoi(u.Uid)
}

func LookupUserStruct(id string) (*user.User, error) {
	var u *user.User

	if _, err := strconv.Atoi(id); err == nil {
		u, err = user.LookupId(id)
		if err != nil {
			return nil, err
		}
	} else {
		u, err = user.Lookup(id)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}

func ShellEscape(s string) string {
	return "'" + strings.Replace(s, "'", "'\"'\"'", -1) + "'"
}

func EnvKeyEscape(s string) string {
	return strings.Replace(strings.Replace(s, "-", "_", -1), ".", "_", -1)
}

func expandEnvironToStringSlice(commandAndArgs, environ []string) []string {
	expanded := []string{}

	items := make(map[string]string)
	for _, val := range environ {
		splits := strings.SplitN(val, "=", 2)
		key := splits[0]
		value := splits[1]
		items[key] = value
	}

	for _, value := range commandAndArgs {
		expanded = append(expanded, os.Expand(value, func(s string) string {
			if v, ok := items[s]; ok {
				return v
			} else {
				return ""
			}
		}))
	}

	return expanded
}

func expandEnvironToString(value string, environ []string) string {
	items := make(map[string]string)
	for _, val := range environ {
		splits := strings.SplitN(val, "=", 2)
		key := splits[0]
		value := splits[1]
		items[key] = value
	}

	return os.Expand(value, func(s string) string {
		if v, ok := items[s]; ok {
			return v
		} else {
			return ""
		}
	})
}

func checkAuthKey(a *Agent, ctx ssh.Context, key ssh.PublicKey) bool {
	config := a.Config
	logger := a.Logger
	fnConfig := a.LookupFunction(ctx.User())

	var keysdata []byte

	authorizedKeysFile := config.AuthorizedKeysFile
	if fnConfig != nil && fnConfig.AuthorizedKeysFile != nil {
		// override authorizedKeysFile by fnConfig
		authorizedKeysFile = *fnConfig.AuthorizedKeysFile
	}

	if authorizedKeysFile != "" {
		if _, err := os.Stat(authorizedKeysFile); err == nil {
			data, err := ioutil.ReadFile(config.AuthorizedKeysFile)
			if err != nil {
				logger.Error(err)
				return false
			}

			keysdata = append(keysdata, data...)
			if len(keysdata) != 0 || keysdata[len(keysdata)-1] != '\n' {
				keysdata = append(keysdata, '\n')
			}
		}
	}

	authorizedKeys := config.AuthorizedKeys
	if fnConfig != nil && fnConfig.AuthorizedKeys != nil {
		// override service config
		authorizedKeys = fnConfig.AuthorizedKeys
	}

	for _, s := range authorizedKeys {
		keysdata = append(keysdata, []byte(s+"\n")...)
	}

	logger.Debugf("authorized_keys\n---start---\n%s---end---", string(keysdata))

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
