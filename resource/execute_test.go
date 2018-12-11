package resource

import (
	"bytes"
	"github.com/kohkimakimoto/cofu/cofu"
	"regexp"
	"testing"
)

func TestExecute(t *testing.T) {
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
	app.Logger.SetOutput(stdout)

	if err := app.LoadRecipe(`
execute "echo hogehoge"
	`); err != nil {
		t.Error(err)
	}
	if err := app.Run(); err != nil {
		t.Error(err)
	}

	output := stdout.String()

	if !regexp.MustCompile(`Complete!`).MatchString(output) {
		t.Errorf("unexpected result %v", output)
	}
}
