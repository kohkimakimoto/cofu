package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
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
		r.Attributes["executed"] = true
	}

	return nil
}

func executeSetCurrentAttributes(r *cofu.Resource) error {
	r.CurrentAttributes["executed"] = false

	return nil
}

func executeRunAction(r *cofu.Resource) error {
	logger := r.App.Logger
	ret := r.MustRunCommand(r.GetStringAttribute("command"))

	logger.Debugf("%s\n", ret.Combined.String())

	return nil
}
