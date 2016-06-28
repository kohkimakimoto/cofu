package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"strings"
)

var SoftwarePackage = &cofu.ResourceType{
	Name: "software_package",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"install"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "name",
			DefaultName: true,
			Required:    true,
		},
		&cofu.StringAttribute{
			Name: "version",
		},
		&cofu.StringAttribute{
			Name: "options",
		},
	},
	PreAction:                packagePreAction,
	SetCurrentAttributesFunc: packageSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"install": packageInstallAction,
		"remove":  packageRemoveAction,
	},
}

func packagePreAction(r *cofu.Resource) error {
	switch r.CurrentAction {
	case "install":
		r.Attributes["installed"] = true
	case "remove":
		r.Attributes["installed"] = false
	}

	return nil
}

func packageSetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()

	name := r.GetStringAttribute("name")

	currentInstalled := r.CheckCommand(c.CheckPackageIsInstalled(name, ""))
	r.CurrentAttributes["installed"] = currentInstalled

	if currentInstalled {
		r.CurrentAttributes["version"] = strings.TrimSpace(r.MustRunCommand(c.GetPackageVersion(name, "")).Stdout.String())
	}

	return nil
}

func packageInstallAction(r *cofu.Resource) error {
	c := r.Infra().Command()

	name := r.GetStringAttribute("name")
	version := r.GetStringAttribute("version")
	options := r.GetStringAttribute("options")

	if !r.CheckCommand(c.CheckPackageIsInstalled(name, version)) {
		r.MustRunCommand(c.InstallPackage(name, version, options))
	}

	return nil
}

func packageRemoveAction(r *cofu.Resource) error {
	c := r.Infra().Command()

	name := r.GetStringAttribute("name")
	options := r.GetStringAttribute("options")

	if r.CheckCommand(c.CheckPackageIsInstalled(name, "")) {
		r.MustRunCommand(c.RemovePackage(name, options))
	}

	return nil
}
