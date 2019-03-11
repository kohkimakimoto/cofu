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
	},
	PreAction:                resourcePreAction,
	SetCurrentAttributesFunc: resourceSetCurrentAttributes,
	ShowDifferences:          resourceShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"run": resourceRunAction,
	},
	UseFallbackAttributes: true,
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
	if !strings.HasSuffix(path, ".lua") {
		path += ".lua"
	}

	// If the path is relative, try to find under the 'resources'
	if !filepath.IsAbs(path) {
		one := filepath.Join(r.Basepath, "resources", path)
		if _, err := os.Stat(one); err == nil {
			path = one
		} else {
			// or under the current directory.
			current := cofu.CurrentDir(r.App.LState)
			path = filepath.Join(current, path)
		}
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
	variables := r.FallbackAttributes
	if variables == nil {
		variables = map[string]interface{}{}
	}

	app := cofu.NewApp()
	defer app.Close()

	app.Logger = r.App.Logger
	app.ResourceTypes = r.App.ResourceTypes
	app.Parent = r.App
	app.Tmpdir = r.App.Tmpdir
	app.Level = r.App.Level + 1
	app.Logger.SetHeader(r.App.LogHeaderWitoutIndent + cofu.GenLogIndent(app.Level))
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

	if err := app.Run(app.DryRun); err != nil {
		return err
	}

	return nil
}

var DefaultBuiltinRecipes = map[string]string{
	//"testing": `print("hello "..var.name)`,
}
