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

func (m *SessionManager) SetSession(sess *Session) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task := sess.TaskConfig
	sessMapPerTask, ok := m.Sessions[task.Name]
	if !ok {
		sessMapPerTask = map[uint64]*Session{}
	}

	max := task.MaxProcesses
	if max > 0 && len(sessMapPerTask) >= max {
		return fmt.Errorf("Limit of max_processes: %d", max)
	}

	sessMapPerTask[sess.ID] = sess
	m.Sessions[task.Name] = sessMapPerTask

	return nil
}

func (m *SessionManager) RemoveSession(sess *Session) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	task := sess.TaskConfig
	if task == nil {
		return
	}

	sessMapPerTask, ok := m.Sessions[task.Name]
	if !ok {
		return
	}

	delete(sessMapPerTask, sess.ID)
	m.Sessions[task.Name] = sessMapPerTask
}

type Session struct {
	ssh.Session
	ID                   uint64
	Agt                  *Agent
	Uid                  int
	Gid                  int
	ForwardAgentListener net.Listener
	TaskConfig           *TaskConfig
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
	task := sess.TaskConfig
	if task == nil {
		return
	}

	if task.KeepSandboxes == 0 {
		return
	}

	if !task.Sandbox {
		return
	}

	sandboxesDir := agt.SandBoxesServiceDir(task.Name)
	files, err := ioutil.ReadDir(sandboxesDir)
	if err != nil {
		logger.Error(err)
	}

	count := len(files)
	keeps := task.KeepSandboxes
	removes := 0
	if keeps > 0 {
		removes = count - keeps
		if removes < 0 {
			removes = 0
		}
	}

	logger.Debugf("%s sandbox(es): %d", task.Name, count)
	logger.Debugf("%s keeps: %d", task.Name, keeps)
	logger.Debugf("%s removes: %d", task.Name, removes)

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
