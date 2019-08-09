package agent

import (
	"encoding/json"
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type SessionManager struct {
	Agt      *Agent
	Sessions map[uint64]*Session
	mutex    *sync.Mutex
}

func NewSessionManager(a *Agent) *SessionManager {
	return &SessionManager{
		Agt:      a,
		Sessions: map[uint64]*Session{},
		mutex:    new(sync.Mutex),
	}
}

func (m *SessionManager) SetSession(sess *Session) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Sessions[sess.ID] = sess

	return len(m.Sessions)
}

func (m *SessionManager) HasSession(sessionID uint64) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.Sessions[sessionID]
	return ok
}

func (m *SessionManager) IsActiveSandbox(sandboxName string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, session := range m.Sessions {
		if session.SandboxName == sandboxName {
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
	ID                   uint64
	SandboxName          string
	Agt                  *Agent
	Uid                  int
	Gid                  int
	ForwardAgentListener net.Listener
}

func NewSession(a *Agent, sshSession ssh.Session) (*Session, error) {
	id, err := a.Gen.NextID()
	if err != nil {
		return nil, err
	}

	sandboxName := fmt.Sprintf("%d", id)
	for _, envLine := range sshSession.Environ() {
		envItem := strings.SplitN(envLine, "=", 2)
		if len(envItem) == 2 {
			key := envItem[0]
			value := envItem[1]
			if key == "COFU_AGENT_SANDBOX_NAME" {
				sandboxName = value
			}
		}
	}

	sess := &Session{
		Session:     sshSession,
		ID:          id,
		SandboxName: sandboxName,
		Agt:         a,
	}

	return sess, nil
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
		sandboxName := file.Name()

		if m.IsActiveSandbox(sandboxName) {
			logger.Debugf("skipped to delete %s. because this is active sandbox", sandboxName)
			continue
		}

		sandboxPath := filepath.Join(sandboxesDir, sandboxName)
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
