package agent

import (
	"encoding/json"
	"github.com/gliderlabs/ssh"
	"io"
	"fmt"
	"net"
)

type Session struct {
	ssh.Session
	ID                   uint64
	Agt                  *Agent
	Uid                  int
	Gid                  int
	ForwardAgentListener net.Listener
	AgentConfig          *AgentConfig
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
