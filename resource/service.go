package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
)

var Service = &cofu.ResourceType{
	Name: "service",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"nothing"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "name",
			DefaultName: true,
			Required:    true,
		},
		&cofu.StringAttribute{
			Name: "provider",
		},
	},
	PreAction:                servicePreAction,
	SetCurrentAttributesFunc: serviceSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"start":   serviceStartAction,
		"stop":    serviceStopAction,
		"restart": serviceRestartAction,
		"reload":  serviceReloadAction,
		"enable":  serviceEnableAction,
		"disable": serviceDisableAction,
	},
}

func servicePreAction(r *cofu.Resource) error {
	switch r.CurrentAction {
	case "start":
		r.Attributes["running"] = true
	case "restart":
		r.Attributes["running"] = true
		r.Attributes["restarted"] = true
		r.CurrentAttributes["restarted"] = false
	case "reload":
		r.Attributes["running"] = true
		r.Attributes["reloaded"] = true
		r.CurrentAttributes["reloaded"] = false
	case "stop":
		r.Attributes["running"] = false
	case "enable":
		r.Attributes["enabled"] = true
	case "disable":
		r.Attributes["enabled"] = false
	}

	if r.GetStringAttribute("provider") == "" {
		c := r.Infra().Command()
		r.Attributes["provider"] = c.DefaultServiceProvider()
	}
	return nil
}

func serviceSetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	r.CurrentAttributes["running"] = r.CheckCommand(c.CheckServiceIsRunningUnderProvider(provider, name))
	r.CurrentAttributes["enabled"] = r.CheckCommand(c.CheckServiceIsEnabledUnderProvider(provider, name))

	return nil
}

func serviceStartAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	if !r.GetBoolCurrentAttribute("running") {
		r.MustRunCommand(c.StartServiceUnderProvider(provider, name))
	}

	return nil
}

func serviceStopAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	if r.GetBoolCurrentAttribute("running") {
		r.MustRunCommand(c.StopServiceUnderProvider(provider, name))
	}

	return nil
}

func serviceRestartAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	r.MustRunCommand(c.RestartServiceUnderProvider(provider, name))

	return nil
}

func serviceReloadAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	if r.GetBoolCurrentAttribute("running") {
		r.MustRunCommand(c.ReloadServiceUnderProvider(provider, name))
	}

	return nil
}

func serviceEnableAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	if !r.GetBoolCurrentAttribute("enabled") {
		r.MustRunCommand(c.EnableServiceUnderProvider(provider, name))
	}

	return nil
}

func serviceDisableAction(r *cofu.Resource) error {
	c := r.Infra().Command()
	name := r.GetStringAttribute("name")
	provider := r.GetStringAttribute("provider")

	if r.GetBoolCurrentAttribute("enabled") {
		r.MustRunCommand(c.DisableServiceUnderProvider(provider, name))
	}

	return nil
}
