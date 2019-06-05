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
		if a.Config.Agent.HotReload {
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

	logger.Infof("Starting SSH protocol server on %s", a.Config.Agent.Addr)

	// start ssh server
	return ssh.ListenAndServe(a.Config.Agent.Addr, nil, options...)
}

func handleSSHSession(a *Agent, sshSession ssh.Session) error {
	logger := a.Logger

	sess, err := NewSession(a, sshSession)
	if err != nil {
		return err
	}

	ptyReq, winCh, isPty := sess.Pty()

	logger.Debugf("cofu-agent got a ssh command: %v", sess.Command())

	task := a.LookupTask(sess.User())
	if task == nil {
		return fmt.Errorf("Task %s is not found", sess.User())
	}
	sess.TaskConfig = task

	if err := a.SessionManager.SetSession(sess); err != nil {
		return err
	}
	defer sess.Terminate()

	logger.Infof("Allocated session %d", sess.ID)

	logger.Debugf("The selected task is %s", task.Name)
	logger.Debugf("task definition: %v", task)

	svEnviron := append(sess.Environ(),
		fmt.Sprintf("COFU_VERSION=%s", cofu.Version),
		fmt.Sprintf("COFU_SSH_USERNAME=%s", sess.User()),
		fmt.Sprintf("COFU_SESSION_ID=%d", sess.ID),
		fmt.Sprintf("COFU_TASK=%s", task.Name),
		fmt.Sprintf("COFU_COMMAND=%s", cofu.BinPath),
	)

	sessCommand := []string{}
	if len(sess.Command()) > 0 {
		sessCommand = sess.Command()
	} else if task.Command != nil && len(task.Command) > 0 {
		sessCommand = task.Command
	}

	svEnviron = append(svEnviron, fmt.Sprintf("COFU_SESSION_COMMAND=%s", strings.Join(sessCommand, " ")))

	if isPty {
		svEnviron = append(svEnviron, "COFU_PTY=1")
	}

	// setup user and group
	var uid, gid int
	if os.Getuid() == 0 {
		userName := expandEnvironToString(task.User, svEnviron)
		if userName != "" {
			uid, gid, err = getUidAndGidFromUsername(userName)
			if err != nil {
				return err
			}
		} else {
			uid = os.Getuid()
			gid = os.Getgid()
		}

		groupName := expandEnvironToString(task.Group, svEnviron)
		if groupName != "" {
			id, err := LookupGroup(groupName)
			if err != nil {
				return err
			}
			gid = id
		}

		logger.Debugf("The service will run by user: %d, group: %d", uid, gid)
	} else {
		logger.Debug("The cofu is running non root user. The service can be executed only by same user who runs cofu.")
		uid = os.Getuid()
		gid = os.Getgid()
		userName := expandEnvironToString(task.User, svEnviron)
		if userName != "" {
			logger.Warnf("ignore 'user' config. because the cofu is running non root user.")
		}
	}
	sess.Uid = uid
	sess.Gid = gid

	// support forward agent
	svEnviron, err = takeForwardAgentIfRequested(sess, svEnviron)
	if err != nil {
		return err
	}

	// setup sandbox
	var sandBoxDir string
	if task.Sandbox {
		sandBoxDir, err = a.CreateSandBoxDirIfNotExist(sess)
		if err != nil {
			return err
		}
		svEnviron = append(svEnviron, fmt.Sprintf("COFU_SANDBOX_DIR=%s", sandBoxDir))

		if task.SandboxSource != "" {
			// remove empty sandbox directory before downloading source.
			if err := os.RemoveAll(sandBoxDir); err != nil {
				return err
			}

			cmd := exec.Command(cofu.BinPath, "-fetch", task.SandboxSource, sandBoxDir)
			cmd.Env = append(os.Environ(), svEnviron...)
			b, err := cmd.CombinedOutput()
			if b != nil {
				logger.Debugf(string(b))
			}

			if err != nil {
				return errors.Wrapf(err, "fetching process output: %s", string(b))
			}
		}
	}

	for _, v := range task.Environment {
		svEnviron = append(svEnviron, expandEnvironToString(v, svEnviron))
	}

	// setup command
	var commandAndArgs []string
	if task.Entrypoint != nil && len(task.Entrypoint) > 0 {
		commandAndArgs = task.Entrypoint
	} else {
		commandAndArgs = []string{}
	}

	commandAndArgs = append(commandAndArgs, sessCommand...)

	if len(commandAndArgs) == 0 {
		return errors.New("You've successfully authenticated, but it does not support accessing without args.")
	}

	// setup current working directory
	wd := expandEnvironToString(task.Directory, svEnviron)
	if wd == "" && sandBoxDir != "" {
		wd = sandBoxDir
	}

	if wd == "" {
		u, err := LookupUserStruct(fmt.Sprintf("%d", uid))
		if err != nil {
			return err
		}
		wd = u.HomeDir
	}

	command := commandAndArgs[0]
	args := []string{}
	if len(commandAndArgs) >= 2 {
		args = commandAndArgs[1:]
	}

	logger.Debugf("command: %v", commandAndArgs)
	// exec command
	cmd := exec.Command(command, args...)
	cmd.Env = append(os.Environ(), svEnviron...)
	if os.Getuid() == 0 {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
	}
	cmd.Dir = wd

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

		if task.Timeout > 0 {
			done := make(chan error)
			go func() {
				done <- cmd.Wait()
			}()

			select {
			case <-time.After(time.Duration(task.Timeout) * time.Second):
				if err := cmd.Process.Kill(); err != nil {
					return errors.Wrapf(err, "failed to kill: "+err.Error())
				}
				return fmt.Errorf("cofu session timeout. it took time over %d sec.", task.Timeout)
			case err := <-done:
				defer close(done)
				if err != nil {
					return err
				}
				return nil
			}
		} else {
			err := cmd.Wait()
			if err != nil {
				return err
			}
		}
	}

	return nil
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

func getUidAndGidFromUsername(userName string) (int, int, error) {
	u, err := LookupUserStruct(userName)
	if err != nil {
		return -1, -1, err
	}

	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return -1, -1, err
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return -1, -1, err
	}

	return uid, gid, nil
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
