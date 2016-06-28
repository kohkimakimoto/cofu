package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"path/filepath"
)

var RemoteFile = &cofu.ResourceType{
	Name: "remote_file",
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
	PreAction:                remoteFilePreAction,
	SetCurrentAttributesFunc: remoteFileSetCurrentAttributes,
	ShowDifferences:          remoteFileShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"create": remoteFileCreateAction,
		"delete": remoteFileDeleteAction,
	},
}

func remoteFilePreAction(r *cofu.Resource) error {
	if r.Attributes["source"] == nil {
		// set default source
		p := r.GetStringAttribute("path")
		r.Attributes["source"] = filepath.Join(r.Basepath, "files", p)
	}

	return filePreAction(r)
}

func remoteFileSetCurrentAttributes(r *cofu.Resource) error {
	return fileSetCurrentAttributes(r)
}

func remoteFileShowDifferences(r *cofu.Resource) error {
	return fileShowDifferences(r)
}

func remoteFileCreateAction(r *cofu.Resource) error {
	return fileCreateAction(r)
}

func remoteFileDeleteAction(r *cofu.Resource) error {
	return fileDeleteAction(r)
}
