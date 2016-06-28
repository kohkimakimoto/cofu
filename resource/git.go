package resource

import (
	"fmt"
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/cofu/infra/backend"
	"github.com/kohkimakimoto/cofu/infra/util"
	"github.com/kohkimakimoto/loglv"
	"log"
	"strings"
)

var Git = &cofu.ResourceType{
	Name: "git",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"sync"},
			Required: true,
		},
		&cofu.StringAttribute{
			Name:        "destination",
			DefaultName: true,
			Required:    true,
		},
		&cofu.StringAttribute{
			Name:     "repository",
			Required: true,
		},
		&cofu.StringAttribute{
			Name: "revision",
		},
		&cofu.BoolAttribute{
			Name:    "recursive",
			Default: false,
		},
	},
	PreAction:                gitPreAction,
	SetCurrentAttributesFunc: gitSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"sync": gitSyncAction,
	},
}

func gitPreAction(r *cofu.Resource) error {
	r.Values["originFetched"] = false

	switch r.CurrentAction {
	case "sync":
		r.Attributes["exist"] = true
	}

	return nil
}

func gitSetCurrentAttributes(r *cofu.Resource) error {
	c := r.Infra().Command()
	dest := r.GetStringAttribute("destination")

	r.CurrentAttributes["exist"] = r.CheckCommand(c.CheckFileIsDirectory(dest))

	return nil
}

func gitSyncAction(r *cofu.Resource) error {
	recursive := r.GetBoolAttribute("recursive")
	repository := r.GetStringAttribute("repository")
	destination := r.GetStringAttribute("destination")
	revision := r.GetStringAttribute("revision")

	if err := gitEnsureGitAvailable(r); err != nil {
		return err
	}

	newRepository := false

	if gitCheckEmptyDir(r) {
		cmd := "git clone"
		if recursive {
			cmd += " --recursive"
		}
		cmd += " " + repository + " " + destination
		r.MustRunCommand(cmd)
		newRepository = true
	}

	var target string
	if revision != "" {
		target = gitGetRevision(r, revision)
	} else {
		gitFetchOrigin(r)
		target = strings.TrimSpace(gitMustRunCommandInRepo(r, "git ls-remote origin HEAD | cut -f1").Stdout.String())
	}

	if newRepository || target != gitGetRevision(r, "HEAD") {
		// TODO: incorrect implementation?
		gitFetchOrigin(r)
		gitMustRunCommandInRepo(r, "git checkout "+target)
	}

	return nil
}

func gitEnsureGitAvailable(r *cofu.Resource) error {
	if r.RunCommand("which git").ExitStatus != 0 {
		return fmt.Errorf("`git` command is not available. Please install git.")
	}

	return nil
}

func gitCheckEmptyDir(r *cofu.Resource) bool {
	dest := util.ShellEscape(r.GetStringAttribute("destination"))

	return r.RunCommand(`test -z "$(ls -A ` + dest + `)"`).Success()
}

func gitGetRevision(r *cofu.Resource, branch string) string {
	result := gitRunCommandInRepo(r, "git rev-list "+util.ShellEscape(branch))
	if result.ExitStatus != 0 {
		gitFetchOrigin(r)
	}

	str := gitMustRunCommandInRepo(r, "git rev-list "+util.ShellEscape(branch)).Stdout.String()
	return strings.TrimSpace(strings.Split(str, "\n")[0])
}

func gitFetchOrigin(r *cofu.Resource) {
	if r.Values["originFetched"].(bool) {
		return
	}

	r.Values["originFetched"] = true
	gitMustRunCommandInRepo(r, "git fetch origin")
}

func gitMustRunCommandInRepo(r *cofu.Resource, command string) *backend.CommandResult {
	ret := gitRunCommandInRepo(r, command)
	if ret.ExitStatus != 0 {
		panic(ret.Combined.String())
	}

	return ret
}

func gitRunCommandInRepo(r *cofu.Resource, command string) *backend.CommandResult {
	opt := &backend.CommandOption{
		User: r.GetStringAttribute("user"),
		Cwd:  r.GetStringAttribute("destination"),
	}

	i := r.Infra()
	command = i.BuildCommand(command, opt)

	if loglv.IsDebug() {
		log.Printf("    (Debug) command: %s", command)
	}

	return i.RunCommand(command)
}
