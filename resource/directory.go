package resource

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/cofu"
	"strings"
)

var Directory = &cofu.ResourceType{
	Name: "directory",
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
			Name: "mode",
		},
		&cofu.StringAttribute{
			Name: "owner",
		},
		&cofu.StringAttribute{
			Name: "group",
		},
	},
	PreAction:                directoryPreAction,
	SetCurrentAttributesFunc: directorySetCurrentAttributes,
	ShowDifferences:          directoryShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"create": directoryCreateAction,
		"delete": directoryDeleteAction,
	},
}

func directoryPreAction(r *cofu.Resource) error {
	switch r.CurrentAction {
	case "create":
		r.Attributes["exist"] = true
	case "delete":
		r.Attributes["exist"] = false
	}

	return nil
}

func directorySetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	exist := r.CheckCommand(c.CheckFileIsDirectory(path))
	r.CurrentAttributes["exist"] = exist

	if exist {
		r.CurrentAttributes["mode"] = strings.TrimSpace(r.MustRunCommand(c.GetFileMode(path)).Stdout.String())
		r.CurrentAttributes["owner"] = strings.TrimSpace(r.MustRunCommand(c.GetFileOwnerUser(path)).Stdout.String())
		r.CurrentAttributes["group"] = strings.TrimSpace(r.MustRunCommand(c.GetFileOwnerGroup(path)).Stdout.String())
	} else {
		r.CurrentAttributes["mode"] = ""
		r.CurrentAttributes["owner"] = ""
		r.CurrentAttributes["group"] = ""
	}

	return nil
}

func directoryShowDifferences(r *cofu.Resource) error {
	if r.GetStringAttribute("mode") != "" {
		r.Attributes["mode"] = fmt.Sprintf("%04s", r.GetStringAttribute("mode"))
	}

	if r.GetStringCurrentAttribute("mode") != "" {
		r.CurrentAttributes["mode"] = fmt.Sprintf("%04s", r.GetStringCurrentAttribute("mode"))
	}

	return cofu.DefaultShowDifferences(r)
}

func directoryCreateAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	mode := r.GetStringAttribute("mode")
	owner := r.GetStringAttribute("owner")
	group := r.GetStringAttribute("group")

	if !r.CheckCommand(c.CheckFileIsDirectory(path)) {
		r.MustRunCommand(c.CreateFileAsDirectory(path))
	}
	if mode != "" {
		r.MustRunCommand(c.ChangeFileMode(path, mode, false))
	}

	if owner != "" || group != "" {
		r.MustRunCommand(c.ChangeFileOwner(path, owner, group, false))
	}
	return nil
}

func directoryDeleteAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	if r.CheckCommand(c.CheckFileIsDirectory(path)) {
		r.MustRunCommand(c.RemoveFile(path))
	}

	return nil
}
