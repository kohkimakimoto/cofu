package resource

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/cofu"
	"path/filepath"
	"strings"
)

var RemoteDirectory = &cofu.ResourceType{
	Name: "remote_directory",
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
			Name:     "source",
			Required: true,
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
	PreAction:                remoteDirectoryPreAction,
	SetCurrentAttributesFunc: remoteDirectorySetCurrentAttributes,
	ShowDifferences:          remoteDirectoryShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"create": remoteDirectoryCreateAction,
		"delete": remoteDirectoryDeleteAction,
	},
}

func remoteDirectoryPreAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")
	exist := r.CheckCommand(c.CheckFileIsDirectory(path))
	r.CurrentAttributes["exist"] = exist

	directory := r.GetStringAttribute("source")
	if !filepath.IsAbs(directory) {
		directory = filepath.Join(r.Basepath, directory)
	}

	temppath, err := r.SendDirectoryToTempDirectory(directory)
	if err != nil {
		return err
	}
	r.Values["temppath"] = temppath

	compareTo := path

	r.Values["compareTo"] = compareTo
	if exist && r.CurrentAction != "delete" {
		r.Attributes["modified"] = r.IsDifferentFilesRecursively(compareTo, temppath)
	} else {
		r.Attributes["modified"] = false
	}

	switch r.CurrentAction {
	case "create":
		r.Attributes["exist"] = true
	case "delete":
		r.Attributes["exist"] = false
	}

	return nil
}

func remoteDirectorySetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	r.CurrentAttributes["modified"] = false
	if r.GetBoolCurrentAttribute("exist") {
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

func remoteDirectoryShowDifferences(r *cofu.Resource) error {
	if r.GetStringAttribute("mode") != "" {
		r.Attributes["mode"] = fmt.Sprintf("%04s", r.GetStringAttribute("mode"))
	}

	if r.GetStringCurrentAttribute("mode") != "" {
		r.CurrentAttributes["mode"] = fmt.Sprintf("%04s", r.GetStringCurrentAttribute("mode"))
	}

	err := cofu.DefaultShowDifferences(r)
	if err != nil {
		return err
	}

	modified := r.GetBoolAttribute("modified")

	if r.CurrentAction != "delete" && modified {
		r.ShowContentDiffRecursively(r.Values["compareTo"].(string), r.Values["temppath"].(string))
	}

	return nil
}

func remoteDirectoryCreateAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	mode := r.GetStringAttribute("mode")
	owner := r.GetStringAttribute("owner")
	group := r.GetStringAttribute("group")

	temppath := r.Values["temppath"]
	changeTarget := temppath.(string)

	if mode != "" {
		r.MustRunCommand(c.ChangeFileMode(changeTarget, mode, false))
	}

	if owner != "" || group != "" {
		r.MustRunCommand(c.ChangeFileOwner(changeTarget, owner, group, false))
	}

	r.MustRunCommand(c.RemoveFile(path))
	r.MustRunCommand(c.MoveFile(temppath.(string), path))

	return nil
}

func remoteDirectoryDeleteAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	if r.CheckCommand(c.CheckFileIsDirectory(path)) {
		r.MustRunCommand(c.RemoveFile(path))
	}

	return nil
}
