package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/loglv"
	"log"
)

var Execute = &cofu.ResourceType{
	Name: "execute",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"run"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "command",
			DefaultName: true,
			Required:    true,
		},
	},
	PreAction:                executePreAction,
	SetCurrentAttributesFunc: executeSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"run": executeRunAction,
	},
}

func executePreAction(r *cofu.Resource) error {
	if r.CurrentAction == "run" {
		r.CurrentAttributes["executed"] = true
	}

	return nil
}

func executeSetCurrentAttributes(r *cofu.Resource) error {
	r.Attributes["executed"] = false

	return nil
}

func executeRunAction(r *cofu.Resource) error {
	ret := r.MustRunCommand(r.GetStringAttribute("command"))

	if loglv.IsDebug() {
		log.Printf("%s\n", ret.Combined.String())
	}

	return nil
}
