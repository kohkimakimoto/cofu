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

	sess, err := NewSession(a, sshSession)
	if err != nil {
		return err
	}

	ptyReq, winCh, isPty := sess.Pty()

	logger.Debugf("cofu-agent got a ssh command: %v", sess.Command())

	numSess := a.SessionManager.SetSession(sess)
	defer sess.Terminate()

	logger.Infof("Allocated session %d", sess.ID)
	logger.Debugf("The session user is %s", sess.User())

	svEnviron := append(sess.Environ(),
		fmt.Sprintf("COFU_AGENT_VERSION=%s", cofu.Version),
		fmt.Sprintf("COFU_AGENT_SESSION_USER=%s", sess.User()),
		fmt.Sprintf("COFU_AGENT_SESSION_ID=%d", sess.ID),
		fmt.Sprintf("COFU_AGENT_SESSION_NUM=%d", numSess),
		fmt.Sprintf("COFU_COMMAND=%s", cofu.BinPath),
	)

	commandAndArgs := []string{}
	if len(sess.Command()) > 0 {
		commandAndArgs = sess.Command()
	} else {
		commandAndArgs = []string{"/bin/bash"}
	}

	svEnviron = append(svEnviron, fmt.Sprintf("COFU_AGENT_SESSION_COMMAND=%s", strings.Join(commandAndArgs, " ")))

	if isPty {
		svEnviron = append(svEnviron, "COFU_AGENT_PTY=1")
	}

	// setup user and group
	uid, gid, err := getUidAndGidFromUsername(sess.User())
	if err != nil {
		return err
	}
	logger.Debugf("This session runs by user: %d, group: %d", uid, gid)

	if os.Getuid() != 0 && os.Getuid() != uid {
		return errors.New("The cofu is running non root user. The connection can be accepted only by same user who runs cofu.")
	}

	sess.Uid = uid
	sess.Gid = gid

	// support forward agent
	svEnviron, err = takeForwardAgentIfRequested(sess, svEnviron)
	if err != nil {
		return err
	}

	// setup sandbox
	sandBoxDir, err := a.CreateSandBoxDirIfNotExist(sess)
	if err != nil {
		return err
	}
	svEnviron = append(svEnviron, fmt.Sprintf("COFU_AGENT_SANDBOX_DIR=%s", sandBoxDir))

	// setup environment variables
	for _, v := range a.Config.Environment {
		svEnviron = append(svEnviron, expandEnvironToString(v, svEnviron))
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
	}

	return cmd.Wait()
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

var SystemAuthorizedKeysFile = "/etc/cofu-agent/authorized_keys"

func checkAuthKey(a *Agent, ctx ssh.Context, key ssh.PublicKey) bool {
	config := a.Config
	logger := a.Logger

	var keysdata []byte

	if SystemAuthorizedKeysFile != config.AuthorizedKeysFile {
		if _, err := os.Stat(SystemAuthorizedKeysFile); err == nil {
			data, err := ioutil.ReadFile(SystemAuthorizedKeysFile)
			if err != nil {
				logger.Error(err)
				return false
			}

			keysdata = append(keysdata, data...)
			if len(keysdata) != 0 && keysdata[len(keysdata)-1] != '\n' {
				keysdata = append(keysdata, '\n')
			}
		}
	}

	if config.AuthorizedKeysFile != "" {
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

	for _, s := range config.AuthorizedKeys {
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
