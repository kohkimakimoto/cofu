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

	sessMapPerService, ok := m.Sessions[sess.User()]
	if !ok {
		sessMapPerService = map[uint64]*Session{}
	}

	//max := srv.MaxProcesses
	//if max > 0 && len(sessMapPerService) >= max {
	//	return fmt.Errorf("Limit of max_processes: %d", max)
	//}

	sessMapPerService[sess.ID] = sess
	m.Sessions[sess.User()] = sessMapPerService

	return len(sessMapPerService)
}

func (m *SessionManager) RemoveSession(sess *Session) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	sessMapPerService, ok := m.Sessions[sess.User()]
	if !ok {
		return
	}

	delete(sessMapPerService, sess.ID)
	m.Sessions[sess.User()] = sessMapPerService
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
	sess.RemoveSandboxes()
}

func (sess *Session) RemoveSandboxes() {
	agt := sess.Agt
	logger := agt.Logger

	if agt.Config.Agent.KeepSandboxes == 0 {
		return
	}

	sandboxesDir := agt.SandBoxesUserDir(sess.User())
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

	logger.Debugf("%s sandbox(es): %d", sess.User(), count)
	logger.Debugf("%s keeps: %d", sess.User(), keeps)
	logger.Debugf("%s removes: %d", sess.User(), removes)

	for i := 0; i < removes; i++ {
		file := files[i]
		sandboxPath := filepath.Join(sandboxesDir, file.Name())
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
