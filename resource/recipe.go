package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"path/filepath"
	"strings"
)

var Recipe = &cofu.ResourceType{
	Name: "recipe",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"run"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "path",
			DefaultName: true,
			Required:    true,
		},
		&cofu.MapAttribute{
			Name:    "variables",
			Default: map[string]interface{}{},
		},
	},
	PreAction:                recipePreAction,
	SetCurrentAttributesFunc: recipeSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"run": recipeRunAction,
	},
}

func recipePreAction(r *cofu.Resource) error {
	if r.CurrentAction == "run" {
		r.Attributes["loaded"] = true
	}

	return nil
}

func recipeSetCurrentAttributes(r *cofu.Resource) error {
	r.CurrentAttributes["loaded"] = false

	return nil
}

func recipeRunAction(r *cofu.Resource) error {
	path := r.GetStringAttribute("path")
	if !filepath.IsAbs(path) {
		current := cofu.CurrentDir(r.App.LState)
		path = filepath.Join(current, path)
	}

	if !strings.HasSuffix(path, ".lua") {
		path += ".lua"
	}

	// load variables
	variables := r.GetMapAttribute("variables")
	if variables == nil {
		variables = map[string]interface{}{}
	}

	app := cofu.NewApp()
	defer app.Close()

	app.LogLevel = r.App.LogLevel
	app.DryRun = r.App.DryRun
	app.ResourceTypes = r.App.ResourceTypes
	app.Parent = r.App
	app.Level = r.App.Level + 1
	app.LogIndent = cofu.LogIndent(app.Level)

	if err := app.Init(); err != nil {
		return err
	}

	if err := app.LoadVariableFromMap(variables); err != nil {
		return err
	}

	if err := app.LoadRecipeFile(path); err != nil {
		return err
	}

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}
