package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"strings"
)

var Link = &cofu.ResourceType{
	Name: "link",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"create"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "link",
			DefaultName: true,
			Required:    true,
		},
		&cofu.StringAttribute{
			Name:     "to",
			Required: true,
		},
		&cofu.BoolAttribute{
			Name: "force",
		},
	},
	PreAction:                linkPreAction,
	SetCurrentAttributesFunc: linkSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"create": linkCreateAction,
	},
}

func linkPreAction(r *cofu.Resource) error {
	switch r.CurrentAction {
	case "create":
		r.Attributes["exist"] = true
	}

	return nil
}

func linkSetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	link := r.GetStringAttribute("link")

	exist := r.CheckCommand(c.CheckFileIsLink(link))
	r.CurrentAttributes["exist"] = exist

	if exist {
		r.CurrentAttributes["to"] = strings.TrimSpace(r.MustRunCommand(c.GetFileLinkTarget(link)).Stdout.String())
	}

	return nil
}

func linkCreateAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	link := r.GetStringAttribute("link")
	to := r.GetStringAttribute("to")
	force := r.GetBoolAttribute("force")

	if !r.CheckCommand(c.CheckFileIsLinkedTo(link, to)) {
		r.MustRunCommand(c.LinkFileTo(link, to, force))
	}

	return nil
}
