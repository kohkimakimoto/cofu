package resource

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/cofu"
	"strings"
)

var File = &cofu.ResourceType{
	Name: "file",
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
			Name: "mode",
		},
		&cofu.StringAttribute{
			Name: "owner",
		},
		&cofu.StringAttribute{
			Name: "group",
		},
	},
	PreAction:                filePreAction,
	SetCurrentAttributesFunc: fileSetCurrentAttributes,
	ShowDifferences:          fileShowDifferences,
	Actions: map[string]cofu.ResourceAction{
		"create": fileCreateAction,
		"delete": fileDeleteAction,
	},
}

func filePreAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	exist := r.CheckCommand(c.CheckFileIsFile(path))
	r.CurrentAttributes["exist"] = exist

	switch r.CurrentAction {
	case "create":
		r.Attributes["exist"] = true
	case "delete":
		r.Attributes["exist"] = false
		//case "edit":
		//	r.Attributes["exist"] = true
		//	// TODO: edit mode
	}

	if r.Attributes["content"] != nil || r.Attributes["source"] != nil {
		var temppath string

		if r.Attributes["content"] != nil {
			t, err := r.SendContentToTempfile([]byte(r.GetStringAttribute("content")))
			if err != nil {
				return err
			}
			temppath = t
		} else if r.Attributes["source"] != nil {
			// "source" is used "remote_file" resource
			t, err := r.SendFileToTempfile(r.GetStringAttribute("source"))
			if err != nil {
				return err
			}
			temppath = t
		}

		r.Values["temppath"] = temppath

		var compareTo string
		if exist {
			compareTo = path
		} else {
			compareTo = "/dev/null"
		}

		r.Values["compareTo"] = compareTo
		r.Attributes["modified"] = r.IsDifferentFiles(compareTo, temppath)

		if r.CurrentAction == "delete" {
			r.Attributes["modified"] = false
		}

	} else {
		r.Attributes["modified"] = false
	}

	return nil
}

func fileSetCurrentAttributes(r *cofu.Resource) error {
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

func fileShowDifferences(r *cofu.Resource) error {
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
		r.ShowContentDiff(r.Values["compareTo"].(string), r.Values["temppath"].(string))
	}

	return nil
}

func fileCreateAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	mode := r.GetStringAttribute("mode")
	owner := r.GetStringAttribute("owner")
	group := r.GetStringAttribute("group")

	modified := r.GetBoolAttribute("modified")
	currentExist := r.GetBoolCurrentAttribute("exist")
	temppath := r.Values["temppath"]

	if !currentExist && temppath == nil {
		r.MustRunCommand("touch " + path)
	}

	if !currentExist && !modified {
		// empty temp file
		r.MustRunCommand("touch " + path)
	}

	var changeTarget string
	if modified {
		changeTarget = temppath.(string)
	} else {
		changeTarget = path
	}

	if mode != "" {
		r.MustRunCommand(c.ChangeFileMode(changeTarget, mode, false))
	}

	if owner != "" || group != "" {
		r.MustRunCommand(c.ChangeFileOwner(changeTarget, owner, group, false))
	}

	if modified {
		r.MustRunCommand(c.MoveFile(temppath.(string), path))
	}

	return nil
}

func fileEditAction(r *cofu.Resource) error {
	// not implemented.
	// TODO: implementing

	return nil
}

func fileDeleteAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	path := r.GetStringAttribute("path")

	if r.CheckCommand(c.CheckFileIsFile(path)) {
		r.MustRunCommand(c.RemoveFile(path))
	}

	return nil
}
