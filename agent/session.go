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
	"strconv"
	"sync"
)

type SessionManager struct {
	Agt      *Agent
	Sessions map[string]map[uint64]*Session
	mutex    *sync.Mutex
}

func NewSessionManager(a *Agent) *SessionManager {
	return &SessionManager{
		Agt:      a,
		Sessions: map[string]map[uint64]*Session{},
		mutex:    new(sync.Mutex),
	}
}

func (m *SessionManager) SetSession(sess *Session) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessMapPerUsername, ok := m.Sessions[sess.User()]
	if !ok {
		sessMapPerUsername = map[uint64]*Session{}
	}

	sessMapPerUsername[sess.ID] = sess
	m.Sessions[sess.User()] = sessMapPerUsername

	return len(sessMapPerUsername)
}

func (m *SessionManager) HasSession(username string, sessionID uint64) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessMapPerUsername, ok := m.Sessions[username]
	if !ok {
		return false
	}

	_, ok = sessMapPerUsername[sessionID]
	return ok
}

func (m *SessionManager) RemoveSession(sess *Session) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessMapPerUsername, ok := m.Sessions[sess.User()]
	if !ok {
		return
	}

	delete(sessMapPerUsername, sess.ID)
	m.Sessions[sess.User()] = sessMapPerUsername
}

type Session struct {
	ssh.Session
	ID                   uint64
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

	sess := &Session{
		Session: sshSession,
		ID:      id,
		Agt:     a,
	}

	return sess, nil
}

func (sess *Session) Terminate() {
	if sess.ForwardAgentListener != nil {
		sess.ForwardAgentListener.Close()
	}

	sess.Agt.SessionManager.RemoveSession(sess)
	sess.Agt.SessionManager.RemoveOldSandboxes(sess.User())
}

func (m *SessionManager) RemoveOldSandboxes(username string) {
	agt := m.Agt
	logger := agt.Logger

	if agt.Config.Agent.KeepSandboxes == 0 {
		return
	}

	sandboxesDir := agt.SandBoxesUserDir(username)
	files, err := ioutil.ReadDir(sandboxesDir)
	if err != nil {
		logger.Error(err)
	}

	count := len(files)
	keeps := agt.Config.Agent.KeepSandboxes
	removes := 0
	if keeps > 0 {
		removes = count - keeps
		if removes < 0 {
			removes = 0
		}
	}

	logger.Debugf("session %s sandbox(es): %d", username, count)
	logger.Debugf("session %s keeps: %d", username, keeps)
	logger.Debugf("session %s removes: %d", username, removes)

	for i := 0; i < removes; i++ {
		file := files[i]
		fileName := file.Name()
		sessionId, err := strconv.ParseUint(fileName, 10, 64)
		if err != nil {
			logger.Error(err)
		}

		if m.HasSession(username, sessionId) {
			logger.Debugf("skipped %v. because this is active session", fileName)
			continue
		}

		sandboxPath := filepath.Join(sandboxesDir, fileName)
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
