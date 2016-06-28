package resource

import (
	"bytes"
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/cofu/gluamapper"
	"github.com/kohkimakimoto/loglv"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

var Template = &cofu.ResourceType{
	Name: "template",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"create"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "path",
			DefaultName: true,
			Required:    true,
		},
		&cofu.StringAttribute{
			Name: "content",
		},
		&cofu.StringAttribute{
			Name: "source",
		},
		&cofu.MapAttribute{
			Name:    "variables",
			Default: map[string]interface{}{},
		},
		&cofu.StringAttribute{
			Name: "mode",
		},
		&cofu.StringAttribute{
			Name: "owner",
		},
		&cofu.StringAttribute{
			Name: "group",
		},
	},
	PreAction:                templatePreAction,
	SetCurrentAttributesFunc: templateSetCurrentAttributes,
	ShowDifferences:          templateShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"create": templateCreateAction,
		"delete": templateDeleteAction,
	},
}

func templatePreAction(r *cofu.Resource) error {
	var templateContent string

	if r.Attributes["content"] != nil {
		templateContent = r.GetStringAttribute("content")
	} else {
		if r.Attributes["source"] == nil {
			// try to load default source
			p := r.GetStringAttribute("path")
			//r.Attributes["source"] = filepath.Join(r.Basepath, "templates", p)
			p = filepath.Join(r.Basepath, "templates", p)

			for _, ext := range []string{".tmpl", ""} {
				ps := p + ext
				r.Attributes["source"] = ps
				if _, err := os.Stat(ps); err == nil {
					if loglv.IsDebug() {
						log.Printf("    (Debug) '%s' is used as template file", ps)
					}
					break
				}
			}
		}

		b, err := ioutil.ReadFile(r.GetStringAttribute("source"))
		if err != nil {
			return err
		}

		templateContent = string(b)
	}

	tmpl, err := template.New("T").Parse(templateContent)
	if err != nil {
		return err
	}

	// load variables
	variables := r.GetMapAttribute("variables")
	if variables == nil {
		variables = map[string]interface{}{}
	}

	// load global 'var' variable
	gVar := map[string]interface{}{}
	gVartb := r.App.LState.GetGlobal("var").(*lua.LTable)
	gluamapper.NewMapper(gluamapper.Option{
		NameFunc: func(s string) string {
			return s
		},
	}).Map(gVartb, &gVar)
	variables["var"] = gVar

	var b bytes.Buffer
	err = tmpl.Execute(&b, variables)
	if err != nil {
		return err
	}

	r.Attributes["content"] = b.String()

	return filePreAction(r)
}

func templateSetCurrentAttributes(r *cofu.Resource) error {
	return fileSetCurrentAttributes(r)
}

func templateShowDifferences(r *cofu.Resource) error {
	return fileShowDifferences(r)
}

func templateCreateAction(r *cofu.Resource) error {
	return fileCreateAction(r)
}

func templateDeleteAction(r *cofu.Resource) error {
	return fileDeleteAction(r)
}
