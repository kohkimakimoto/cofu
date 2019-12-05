package agent

import (
	"encoding/json"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/jehiah/go-strftime"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type SessionManager struct {
	Agt      *Agent
	Sessions map[string]*Session
	mutex    *sync.Mutex
}

func NewSessionManager(a *Agent) *SessionManager {
	return &SessionManager{
		Agt:      a,
		Sessions: map[string]*Session{},
		mutex:    new(sync.Mutex),
	}
}

func (m *SessionManager) SetSession(sess *Session) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Sessions[sess.ID] = sess

	return len(m.Sessions)
}

func (m *SessionManager) HasSession(sessionID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.Sessions[sessionID]
	return ok
}

func (m *SessionManager) IsActiveSandbox(sandbox string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, session := range m.Sessions {
		if session.Sandbox == sandbox {
			return true
		}
	}

	return false
}

func (m *SessionManager) RemoveSession(sess *Session) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.Sessions, sess.ID)
}

type Session struct {
	ssh.Session
	ID                   string
	Sandbox              string
	Agt                  *Agent
	Uid                  int
	Gid                  int
	UserStruct           *user.User
	ForwardAgentListener net.Listener
}

func NewSession(a *Agent, sshSession ssh.Session) *Session {
	id := generateSessionID()

	sandboxName := id
	for _, envLine := range sshSession.Environ() {
		envItem := strings.SplitN(envLine, "=", 2)
		if len(envItem) == 2 {
			key := envItem[0]
			value := envItem[1]
			if key == "COFU_AGENT_SANDBOX" {
				sandboxName = value
			}
		}
	}

	return &Session{
		Session: sshSession,
		ID:      id,
		Sandbox: sandboxName,
		Agt:     a,
	}
}

func (sess *Session) Terminate() {
	if sess.ForwardAgentListener != nil {
		sess.ForwardAgentListener.Close()
	}

	sess.Agt.SessionManager.RemoveSession(sess)
	sess.Agt.SessionManager.RemoveOldSandboxes()
}

func (m *SessionManager) RemoveOldSandboxes() {
	agt := m.Agt
	logger := agt.Logger

	if agt.Config.KeepSandboxes == 0 {
		// does not clean up sandboxes automatically.
		return
	}

	sandboxesDir := agt.Config.SandboxesDirectory
	files, err := ioutil.ReadDir(sandboxesDir)
	if err != nil {
		logger.Error(err)
	}

	count := len(files)
	keeps := agt.Config.KeepSandboxes
	removes := 0
	if keeps > 0 {
		removes = count - keeps
		if removes < 0 {
			removes = 0
		}
	}

	logger.Debugf("sandbox(es): %d", count)
	logger.Debugf("keeps: %d", keeps)
	logger.Debugf("removes: %d", removes)

	for i := 0; i < removes; i++ {
		file := files[i]
		sandbox := file.Name()

		if m.IsActiveSandbox(sandbox) {
			logger.Debugf("skipped to delete %s. because this is active sandbox", sandbox)
			continue
		}

		sandboxPath := filepath.Join(sandboxesDir, sandbox)
		if err := os.RemoveAll(sandboxPath); err != nil {
			logger.Error(err)
		}

		logger.Debugf("deleted %v", sandboxPath)
	}
}

func (sess *Session) First() string {
	command := sess.Command()

	first := ""
	if len(command) > 0 {
		first = command[0]
	}
	return first
}

func (sess *Session) JSON(i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	//_, err = sess.Write(b)
	_, err = io.WriteString(sess, fmt.Sprintf("%s\n", string(b)))
	if err != nil {
		return err
	}

	return nil
}

func (sess *Session) JSONPretty(i interface{}) error {
	b, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return err
	}
	_, err = io.WriteString(sess, fmt.Sprintf("%s\n", string(b)))
	if err != nil {
		return err
	}

	return nil
}

func (sess *Session) String(in string) error {
	_, err := io.WriteString(sess, in)
	if err != nil {
		return err
	}

	return nil
}

func (sess *Session) Stringf(format string, a ...interface{}) error {
	return sess.String(fmt.Sprintf(format, a...))
}

var (
	idLastTimestamp string
	idLock          sync.Mutex
)

func generateSessionID() string {
	idLock.Lock()
	defer idLock.Unlock()

	timestamp := nextTimestamp()

	for idLastTimestamp == timestamp {
		time.Sleep(1 * time.Millisecond)
		timestamp = nextTimestamp()
	}

	idLastTimestamp = timestamp
	return timestamp
}

func nextTimestamp() string {
	t := time.Now()
	timestamp := strftime.Format("%Y%m%d%H%M%S", t)
	milli := fmt.Sprintf("%03d", (t.UnixNano()/1e6)%1000)
	return timestamp + milli
}
