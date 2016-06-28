package cofu

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/cofu/resource"
	"os"
	"testing"
)

func TestIntegration(t *testing.T) {
	if os.Getenv("COFU_INTEGRATION_TEST") == "" {
		return
	}

	app := newApp()
	defer app.Close()

	// At now, just runs...
	app.LoadRecipeFile("_tests/recipe.lua")
	if err := app.Run(); err != nil {
		t.Error(err)
	}
}

func newApp() *cofu.App {
	app := cofu.NewApp()
	app.ResourceTypes = []*cofu.ResourceType{
		resource.Directory,
		resource.Execute,
		resource.File,
		resource.Git,
		resource.Group,
		resource.Link,
		resource.LuaFunction,
		resource.SoftwarePackage,
		resource.Service,
		resource.RemoteFile,
		resource.Template,
		resource.User,
	}
	app.LogLevel = "debug"

	if err := app.Init(); err != nil {
		panic(err)
	}

	return app
}
