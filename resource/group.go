package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"strconv"
	"strings"
)

var Group = &cofu.ResourceType{
	Name: "group",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"create"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "groupname",
			DefaultName: true,
			Required:    true,
		},
		&cofu.IntegerAttribute{
			Name: "gid",
		},
	},
	PreAction:                groupPreAction,
	SetCurrentAttributesFunc: groupSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"create": groupCreateAction,
	},
}

func groupPreAction(r *cofu.Resource) error {
	r.Attributes["exist"] = true

	return nil
}

func groupSetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	groupname := r.GetStringAttribute("groupname")

	exist := r.CheckCommand(c.CheckGroupExists(groupname))
	r.CurrentAttributes["exist"] = exist

	if exist {
		i, err := strconv.Atoi(strings.TrimSpace(r.MustRunCommand(c.GetGroupGid(groupname)).Stdout.String()))
		if err != nil {
			return err
		}

		r.CurrentAttributes["gid"] = &cofu.Integer{
			V:     i,
			IsNil: false,
		}
	}

	return nil
}

func groupCreateAction(r *cofu.Resource) error {
	c := r.Infra().Command()

	groupname := r.GetStringAttribute("groupname")
	gid := r.GetIntegerAttribute("gid")

	exist := r.CheckCommand(c.CheckGroupExists(groupname))
	if exist {
		currentGid := r.GetIntegerCurrentAttribute("gid")
		if gid != nil && !gid.Nil() && gid.String() != currentGid.String() {
			// group exists and modify gid
			r.MustRunCommand(c.UpdateGroupGid(groupname, gid.String()))
		}
	} else {
		options := map[string]string{
			"gid": gid.String(),
		}
		r.MustRunCommand(c.AddGroup(groupname, options))
	}

	return nil
}
