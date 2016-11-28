package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"strings"
)

var User = &cofu.ResourceType{
	Name: "user",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"create"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "username",
			DefaultName: true,
			Required:    true,
		},
		&cofu.StringAttribute{
			Name: "gid",
		},
		&cofu.StringAttribute{
			Name: "home",
		},
		&cofu.StringAttribute{
			Name: "password",
		},
		&cofu.BoolAttribute{
			Name:    "system_user",
			Default: false,
		},
		&cofu.IntegerAttribute{
			Name: "uid",
		},
		&cofu.StringAttribute{
			Name: "shell",
		},
		&cofu.BoolAttribute{
			Name:    "create_home",
			Default: false,
		},
	},
	PreAction:                userPreAction,
	SetCurrentAttributesFunc: userSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"create": userCreateAction,
	},
}

func userPreAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	r.Attributes["exist"] = true

	gid := r.GetStringAttribute("gid")
	if gid != "" {
		r.Attributes["gid"] = strings.TrimSpace(r.MustRunCommand(c.GetGroupGid(gid)).Stdout.String())
	}

	return nil
}

func userSetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	username := r.GetStringAttribute("username")

	exist := r.CheckCommand(c.CheckUserExists(username))
	r.CurrentAttributes["exist"] = exist

	if exist {
		r.CurrentAttributes["uid"] = strings.TrimSpace(r.MustRunCommand(c.GetUserUid(username)).Stdout.String())
		r.CurrentAttributes["gid"] = strings.TrimSpace(r.MustRunCommand(c.GetUserGid(username)).Stdout.String())
		r.CurrentAttributes["home"] = strings.TrimSpace(r.MustRunCommand(c.GetUserHomeDirectory(username)).Stdout.String())
		r.CurrentAttributes["shell"] = strings.TrimSpace(r.MustRunCommand(c.GetUserLoginShell(username)).Stdout.String())

		result := r.RunCommand(c.GetUserEncryptedPassword(username))
		if result.Success() {
			r.CurrentAttributes["password"] = strings.TrimSpace(result.Stdout.String())
		}
	}

	return nil
}

func userCreateAction(r *cofu.Resource) error {
	c := r.Infra().Command()

	username := r.GetStringAttribute("username")
	gid := r.GetStringAttribute("gid")
	uid := r.GetIntegerAttribute("uid")
	password := r.GetStringAttribute("password")
	home := r.GetStringAttribute("home")
	shell := r.GetStringAttribute("shell")
	systemUser := r.GetBoolAttribute("system_user")
	createHome := r.GetBoolAttribute("create_home")

	exist := r.CheckCommand(c.CheckUserExists(username))
	if exist {
		currentUid := r.GetStringCurrentAttribute("uid")
		if uid != nil && !uid.Nil() && uid.String() != currentUid {
			r.MustRunCommand(c.UpdateUserUid(username, uid.String()))
		}

		currentGid := r.GetStringCurrentAttribute("gid")
		if gid != "" && gid != currentGid {
			r.MustRunCommand(c.UpdateUserGid(username, gid))
		}

		currentPassword := r.GetStringCurrentAttribute("password")
		if password != "" && password != currentPassword {
			r.MustRunCommand(c.UpdateUserEncryptedPassword(username, password))
		}

		currentHome := r.GetStringCurrentAttribute("home")
		if home != "" && home != currentHome {
			r.MustRunCommand(c.UpdateUserHomeDirectory(username, home))
		}

		currentShell := r.GetStringCurrentAttribute("shell")
		if shell != "" && shell != currentShell {
			r.MustRunCommand(c.UpdateUserLoginShell(username, shell))
		}
	} else {
		system_user := ""
		if systemUser {
			system_user = "true"
		}
		create_home := ""
		if createHome {
			create_home = "true"
		}

		uidString := ""
		if uid != nil && !uid.Nil() {
			uidString = uid.String()
		}

		options := map[string]string{
			"gid":            gid,
			"home_directory": home,
			"password":       password,
			"system_user":    system_user,
			"uid":            uidString,
			"shell":          shell,
			"create_home":    create_home,
		}

		r.MustRunCommand(c.AddUser(username, options))
	}

	return nil
}
