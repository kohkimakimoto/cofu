package cofu

import (
	"testing"
	"os"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TEST") == "" {
		t.Skip()
	}


	t.Log("Starting integration tests")
}