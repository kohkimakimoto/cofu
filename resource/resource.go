package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"os"
	"path/filepath"
	"strings"
)

var Resource = &cofu.ResourceType{
	Name: "resource",
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
	PreAction:                resourcePreAction,
	SetCurrentAttributesFunc: resourceSetCurrentAttributes,
	ShowDifferences:          resourceShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"run": resourceRunAction,
	},
}

func resourcePreAction(r *cofu.Resource) error {
	if r.CurrentAction == "run" {
		r.Attributes["loaded"] = true
	}

	return nil
}

func resourceSetCurrentAttributes(r *cofu.Resource) error {
	r.CurrentAttributes["loaded"] = false

	return nil
}

func resourceShowDifferences(r *cofu.Resource) error {
	return nil
}

func resourceRunAction(r *cofu.Resource) error {
	path := r.GetStringAttribute("path")
	if !filepath.IsAbs(path) {
		current := cofu.CurrentDir(r.App.LState)
		path = filepath.Join(current, path)
	}

	if !strings.HasSuffix(path, ".lua") {
		path += ".lua"
	}

	var builtInRecipe string
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// not found. try to find builtin recipe
		path = r.GetStringAttribute("path")
		recipe, ok := r.App.BuiltinRecipes[path]
		if ok {
			builtInRecipe = recipe
		}
	}

	// load variables
	variables := r.GetMapAttribute("variables")
	if variables == nil {
		variables = map[string]interface{}{}
	}

	app := cofu.NewApp()
	defer app.Close()

	app.Logger = r.App.Logger
	app.DryRun = r.App.DryRun
	app.ResourceTypes = r.App.ResourceTypes
	app.Parent = r.App
	app.Level = r.App.Level + 1
	app.LogIndent = cofu.LogIndent(app.Level)
	app.BuiltinRecipes = r.App.BuiltinRecipes

	if err := app.Init(); err != nil {
		return err
	}

	if err := app.LoadVariableFromMap(variables); err != nil {
		return err
	}

	if builtInRecipe != "" {
		if err := app.LoadRecipe(builtInRecipe); err != nil {
			return err
		}
	} else {
		if err := app.LoadRecipeFile(path); err != nil {
			return err
		}
	}

	if err := app.Run(); err != nil {
		return err
	}

	return nil
}

var DefaultBuiltinRecipes = map[string]string{
	"testing": `print("testing!")`,
}
