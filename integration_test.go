package cofu

import (
	"testing"
	"os"
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/cofu/resource"
	"path/filepath"
)

var testRecipeFiles []string = []string{
	"_tests/resource_directory.lua",
	"_tests/resource_execute.lua",
	"_tests/resource_template.lua",
}

var wd string

func TestIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TEST") == "" {
		t.Skip()
	}

	t.Log("Starting integration tests")

	d, err := os.Getwd()
	if err != nil {
		t.Error(err)
		return
	}

	d, err = filepath.Abs(d)
	if err != nil {
		t.Error(err)
		return
	}

	wd = d

	for _, recipeFile := range testRecipeFiles {
		testRecipeFile(t, recipeFile)
	}
}

func testRecipeFile(t *testing.T, recipeFile string) {
	if err := os.Chdir(wd); err != nil {
		t.Error(err)
		return
	}

	t.Logf("testing %s\n",recipeFile)

	app := cofu.NewApp()
	defer app.Close()

	app.ResourceTypes = resource.ResourceTypes
	if err := app.Init(); err != nil {
		t.Error(err)
	}
	if err := app.LoadRecipeFile(recipeFile); err != nil {
		t.Error(err)
	}
	if err := app.Run(); err != nil {
		t.Error(err)
	}
}