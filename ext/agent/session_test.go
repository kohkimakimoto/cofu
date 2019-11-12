package agent

import "testing"

func TestGenerateSessionID(t *testing.T) {
	id := generateSessionID()

	t.Logf("id: %s", id)
}
