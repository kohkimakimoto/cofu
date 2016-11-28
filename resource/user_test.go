package resource

import (
	"bytes"
	"github.com/kohkimakimoto/cofu/cofu"
	"log"
	"regexp"
	"testing"
)

func TestUserAdd(t *testing.T) {
	// TODO: improve
	// at now, just run
	app := cofu.NewApp()
	defer app.Close()
	app.ResourceTypes = ResourceTypes

	if err := app.Init(); err != nil {
		t.Error(err)
	}

	// override output
	stdout := new(bytes.Buffer)
	log.SetOutput(stdout)

	if err := app.LoadRecipe(`
user "test_user" {
    home = "/home/test_user",
    shell = "/sbin/nologin",
}

	`); err != nil {
		t.Error(err)
	}
	if err := app.Run(); err != nil {
		t.Error(err)
	}

	output := stdout.String()

	t.Log(output)

	if !regexp.MustCompile(`Complete!`).MatchString(output) {
		t.Errorf("unexpected result %v", output)
	}
}
