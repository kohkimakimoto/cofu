package cofu

import (
	"bufio"
	"fmt"
	"github.com/kohkimakimoto/cofu/infra"
	"github.com/kohkimakimoto/cofu/infra/backend"
	"github.com/kohkimakimoto/cofu/infra/util"
	"github.com/kohkimakimoto/cofu/support/color"
	"github.com/kohkimakimoto/loglv"
	"github.com/yuin/gopher-lua"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

type Resource struct {
	Name string
	// Basepath is a directory path that includes recipe file defines this resource.
	Basepath          string
	Attributes        map[string]interface{}
	AttributesLValues map[string]lua.LValue
	CurrentAttributes map[string]interface{}
	Notifications     []*Notification
	ResourceType      *ResourceType
	App               *App
	CurrentAction     string
	Values            map[string]interface{}
	updated           bool
}

func NewResource(name string, resourceType *ResourceType, app *App) *Resource {
	basepath, err := basepath(app.LState)
	if err != nil {
		wd, err2 := os.Getwd()
		if err2 != nil {
			panic(err2)
		}
		basepath = wd
		if loglv.IsDebug() {
			log.Printf("    (Debug) Couldn't get the resource basepath in the lua state (err: %v). so it uses current working directory %s", err, basepath)

		}
	}
	return &Resource{
		Name:              name,
		Attributes:        map[string]interface{}{},
		AttributesLValues: map[string]lua.LValue{},
		CurrentAttributes: map[string]interface{}{},
		ResourceType:      resourceType,
		App:               app,
		Basepath:          basepath,
		Values:            map[string]interface{}{},
	}
}

// Desc returns string like 'resource_type[name]'
func (r *Resource) Desc() string {
	return fmt.Sprintf("%s[%s]", r.ResourceType.Name, r.Name)
}

func (r *Resource) GetRawAttribute(key string) interface{} {
	return r.Attributes[key]
}

func (r *Resource) GetRawCurrentAttribute(key string) interface{} {
	return r.CurrentAttributes[key]
}

func (r *Resource) GetBoolAttribute(key string) bool {
	a, ok := r.Attributes[key]
	if !ok {
		return false
	}

	b, ok := a.(bool)
	if !ok {
		panic(fmt.Sprintf("'%s' is not bool value.", key))
	}

	return b
}

func (r *Resource) GetBoolCurrentAttribute(key string) bool {
	a, ok := r.CurrentAttributes[key]
	if !ok {
		return false
	}

	b, ok := a.(bool)
	if !ok {
		panic(fmt.Sprintf("'%s' is not bool value.", key))
	}

	return b
}

func (r *Resource) GetStringAttribute(key string) string {
	a, ok := r.Attributes[key]
	if !ok {
		return ""
	}

	b, ok := a.(string)
	if !ok {
		panic(fmt.Sprintf("'%s' is not string value.", key))
	}

	return b
}

func (r *Resource) GetStringCurrentAttribute(key string) string {
	a, ok := r.CurrentAttributes[key]
	if !ok {
		return ""
	}

	b, ok := a.(string)
	if !ok {
		panic(fmt.Sprintf("'%s' is not string value.", key))
	}

	return b
}

func (r *Resource) GetStringSliceAttribute(key string) []string {
	a, ok := r.Attributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a string or strings.", key, v))
	}
}

func (r *Resource) GetStringSliceCurrentAttribute(key string) []string {
	a, ok := r.CurrentAttributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a string or strings.", key, v))
	}
}

func (r *Resource) GetMapAttribute(key string) map[string]interface{} {
	a, ok := r.Attributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case map[string]interface{}:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a map of key/value pair.", key, v))
	}
}

func (r *Resource) GetMapCurrentAttribute(key string) map[string]interface{} {
	a, ok := r.CurrentAttributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case map[string]interface{}:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a map of key/value pair.", key, v))
	}
}

func (r *Resource) GetIntegerAttribute(key string) *Integer {
	a, ok := r.Attributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case *Integer:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a number.", key, v))
	}
}

func (r *Resource) GetIntegerCurrentAttribute(key string) *Integer {
	a, ok := r.CurrentAttributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case *Integer:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a number.", key, v))
	}
}

func (r *Resource) GetLFunctionAttribute(key string) *lua.LFunction {
	a, ok := r.Attributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case *lua.LFunction:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a function.", key, v))
	}
}

func (r *Resource) GetLFunctionCurrentAttribute(key string) *lua.LFunction {
	a, ok := r.CurrentAttributes[key]
	if !ok {
		return nil
	}

	switch v := a.(type) {
	case *lua.LFunction:
		return v
	default:
		panic(fmt.Sprintf("'%s' is not supported value type %v. it should be as a function.", key, v))
	}
}

func (r *Resource) Run(specificAction string) error {
	r.updated = false

	var actions []string
	if specificAction != "" {
		actions = []string{specificAction}
	} else {
		actions = r.GetStringSliceAttribute("action")
	}

	if len(actions) == 1 && actions[0] == "nothing" {
		// nothing to do.
		return nil
	}

	log.Print(color.FgBold(fmt.Sprintf("==> Evaluating %s", r.Desc())))

	if loglv.IsDebug() {
		log.Printf("    (Debug) Resource basepath: %s", r.Basepath)
	}

	err := os.Chdir(r.Basepath)
	if err != nil {
		return err
	}

	if loglv.IsDebug() {
		log.Printf("    (Debug) Changed current directory: %s", r.Basepath)
	}

	if r.doNotRunBecauseOfOnlyIf() {
		log.Printf("    Execution skipped because of only_if attribute.")
		return nil
	}

	if r.doNotRunBecauseOfNotIf() {
		log.Printf("    Execution skipped because of not_if attribute.")
		return nil
	}

	for _, action := range actions {
		if err := r.runAction(action); err != nil {
			return err
		}
	}

	if !r.App.DryRun {
		// verify
		if err := r.verify(); err != nil {
			return err
		}
	}

	if r.updated {
		if err := r.notify(); err != nil {
			return err
		}
	}

	r.updated = false

	return nil
}

func (r *Resource) verify() error {
	commands := r.GetStringSliceAttribute("verify")
	if commands == nil {
		return nil
	}

	log.Printf("    Verifying...")
	for _, c := range commands {
		ret := r.RunCommand(c)
		if ret.Failure() {
			return fmt.Errorf("Verifying command '%s' failed with status '%d'. %s", c, ret.ExitStatus, ret.Stderr.String())
		}
	}

	return nil
}

func (r *Resource) notify() error {
	for _, n := range r.Notifications {
		message := fmt.Sprintf("    %s: Notifying %s to %s", r.Desc(), n.Action, n.TargetResourceDesc)
		if n.Delayed() {
			message = fmt.Sprintf("%s (delayed)", message)
		} else if n.Immediately() {
			message = fmt.Sprintf("%s (immediately)", message)
		}

		log.Print(message)

		if n.Delayed() {
			r.App.EnqueueDelayeNotification(n)
		} else if n.Immediately() {
			if err := n.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *Resource) clearCurrentAttributes() {
	r.CurrentAttributes = map[string]interface{}{}
}

func (r *Resource) runAction(action string) error {
	resourceType := r.ResourceType
	actionFunc, ok := resourceType.Actions[action]
	if !ok {
		return fmt.Errorf("Unsupported action '%s'.", action)
	}

	r.CurrentAction = action
	r.clearCurrentAttributes()

	if resourceType.PreAction != nil {
		if loglv.IsDebug() {
			log.Printf("    (Debug) Processing '%s' PreAction", r.Desc())
		}
		err := resourceType.PreAction(r)
		if err != nil {
			return err
		}

		if loglv.IsDebug() {
			log.Printf("    (Debug) Finished '%s' PreAction", r.Desc())
		}
	}

	if resourceType.SetCurrentAttributesFunc != nil {
		if loglv.IsDebug() {
			log.Printf("    (Debug) Processing '%s' SetCurrentAttributes", r.Desc())
		}

		err := resourceType.SetCurrentAttributesFunc(r)
		if err != nil {
			return err
		}

		if loglv.IsDebug() {
			log.Printf("    (Debug) Finished '%s' SetCurrentAttributes", r.Desc())
		}
	}

	if resourceType.ShowDifferences != nil {
		if loglv.IsDebug() {
			log.Printf("    (Debug) Processing '%s' ShowDifferences", r.Desc())
		}

		err := resourceType.ShowDifferences(r)
		if err != nil {
			return err
		}

		if loglv.IsDebug() {
			log.Printf("    (Debug) Finished '%s' ShowDifferences", r.Desc())
		}
	}

	if !r.different() {
		// run action only if the attributes change.
		if loglv.IsDebug() {
			log.Printf("    (Debug) There are not attributes to change '%s'", r.Desc())
		}
		return nil
	}

	if !r.App.DryRun {
		if loglv.IsDebug() {
			log.Printf("    (Debug) Processing '%s' action: '%s'", r.Desc(), action)
		}

		err := actionFunc(r)
		if err != nil {
			return err
		}

		if loglv.IsDebug() {
			log.Printf("    (Debug) Finished '%s' action: '%s'", r.Desc(), action)
		}

		r.Update()
	}

	return nil
}

// different returns true if the resource's attributes different with current attributes.
// see also DefaultShowDifferences
func (r *Resource) different() bool {
	if loglv.IsDebug() {
		log.Printf("    (Debug) Checking difference of '%s'", r.Desc())
	}

	var keys []string
	for key, _ := range r.CurrentAttributes {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	if loglv.IsDebug() {
		log.Printf("    (Debug) Checked keys are '%v'", keys)
	}

	for _, key := range keys {
		if loglv.IsDebug() {
			log.Printf("    (Debug) Checking difference of the key '%s'", key)
		}

		currentValue := r.CurrentAttributes[key]
		if comparable, ok := currentValue.(ComparableValue); ok {
			if comparable.Nil() {
				currentValue = nil
			} else {
				currentValue = comparable.String()
			}
		}

		value := r.Attributes[key]
		if comparable, ok := value.(ComparableValue); ok {
			if comparable.Nil() {
				value = nil
			} else {
				value = comparable.String()
			}
		}

		if loglv.IsDebug() {
			log.Printf("    (Debug) Checking difference '%s' (currentAttr: '%v') => (attr: '%v')", key, currentValue, value)
		}

		if currentValue == nil || value == nil {
			// ignore
		} else if currentValue == nil && value != nil {
			return true
		} else if currentValue == value || value == nil {
			// ignore. not change
		} else {
			if comparable, ok := currentValue.(ComparableValue); ok {
				if comparable.Nil() {
					return false
				}
			}
			return true
		}
	}

	return false
}

func (r *Resource) doNotRunBecauseOfOnlyIf() bool {
	command := r.GetStringAttribute("only_if")
	if command == "" {
		return false
	}

	return r.RunCommand(command).ExitStatus != 0
}

func (r *Resource) doNotRunBecauseOfNotIf() bool {
	command := r.GetStringAttribute("not_if")
	if command == "" {
		return false
	}

	return r.RunCommand(command).ExitStatus == 0
}

func (r *Resource) MustRunCommand(command string) *backend.CommandResult {
	ret := r.RunCommand(command)
	if ret.ExitStatus != 0 {
		panic(ret.Combined.String())
	}

	return ret
}

func (r *Resource) CheckCommand(command string) bool {
	return r.RunCommand(command).ExitStatus == 0
}

func (r *Resource) RunCommand(command string) *backend.CommandResult {
	opt := &backend.CommandOption{
		User: r.GetStringAttribute("user"),
		Cwd:  r.GetStringAttribute("cwd"),
	}

	i := r.Infra()
	command = i.BuildCommand(command, opt)

	if loglv.IsDebug() {
		log.Printf("    (Debug) command: %s", command)
	}

	return i.RunCommand(command)
}

func (r *Resource) SendContentToTempfile(content []byte) (string, error) {
	return r.App.SendContentToTempfile(content)
}

func (r *Resource) SendFileToTempfile(file string) (string, error) {
	return r.App.SendFileToTempfile(file)
}

func (r *Resource) SendDirectoryToTempDirectory(src string) (string, error) {
	return r.App.SendDirectoryToTempDirectory(src)
}

func (r *Resource) IsDifferentFiles(from, to string) bool {
	status := r.RunCommand("diff -q " + util.ShellEscape(from) + " " + util.ShellEscape(to)).ExitStatus
	switch status {
	case 1:
		// diff found
		return true
	case 2:
		panic("diff command exited with 2")
	}

	return false
}

func (r *Resource) IsDifferentFilesRecursively(from, to string) bool {
	status := r.RunCommand("diff -r -q " + util.ShellEscape(from) + " " + util.ShellEscape(to)).ExitStatus
	switch status {
	case 1:
		// diff found
		return true
	case 2:
		panic("diff command exited with 2")
	}

	return false
}

func (r *Resource) ShowContentDiff(from, to string) {
	diff := fmt.Sprintf("diff -u %s %s", util.ShellEscape(from), util.ShellEscape(to))

	if loglv.IsDebug() {
		log.Printf("    (Debug) diff: %s", diff)
	}

	stdout := r.RunCommand(diff).Stdout
	// I intentionally doesn't use bufio.Scanner to prevent bufio.Scanner: token too long
	// see https://github.com/kohkimakimoto/cofu/issues/18
	reader := bufio.NewReader(&stdout)
	for {
		linebytes, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		line := string(linebytes)

		if strings.HasPrefix(line, "+") {
			log.Printf(color.FgG("    %s", line))
		} else if strings.HasPrefix(line, "-") {
			log.Printf(color.FgR("    %s", line))
		} else {
			log.Printf("    %s", line)
		}
	}
}

func (r *Resource) ShowContentDiffRecursively(from, to string) {
	diff := fmt.Sprintf("diff -r -u %s %s", util.ShellEscape(from), util.ShellEscape(to))

	if loglv.IsDebug() {
		log.Printf("    (Debug) diff: %s", diff)
	}

	stdout := r.RunCommand(diff).Stdout
	reader := bufio.NewReader(&stdout)
	for {
		linebytes, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
		line := string(linebytes)

		if strings.HasPrefix(line, "+") {
			log.Printf(color.FgG("    %s", line))
		} else if strings.HasPrefix(line, "-") {
			log.Printf(color.FgR("    %s", line))
		} else {
			log.Printf("    %s", line)
		}
	}
}

func (r *Resource) Infra() *infra.Infra {
	return r.App.Infra
}

func (r *Resource) Update() {
	if r.updated {
		return
	}

	r.updated = true
	if loglv.IsDebug() {
		log.Printf("    (Debug) Resource '%s' is updated.", r.Desc())
	}
}

func (r Resource) IsUpdated() bool {
	return r.updated
}

func DefaultShowDifferences(r *Resource) error {
	// for constant order
	var keys []string
	for key, _ := range r.CurrentAttributes {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		currentValue := r.CurrentAttributes[key]
		if comparable, ok := currentValue.(ComparableValue); ok {
			if comparable.Nil() {
				currentValue = nil
			} else {
				currentValue = comparable.String()
			}
		}

		value := r.Attributes[key]
		if comparable, ok := value.(ComparableValue); ok {
			if comparable.Nil() {
				value = nil
			} else {
				value = comparable.String()
			}
		}

		if currentValue == nil || value == nil {
			// ignore
		} else if currentValue == nil && value != nil {
			log.Printf(color.FgGB("%s: '%s' will be '%v'", r.Desc(), key, value))
		} else if currentValue == value || value == nil {
			// ignore. not change
			if loglv.IsDebug() {
				log.Printf("    (Debug) %s: %s will not change (current value is '%v')", r.Desc(), key, currentValue)
			}
		} else {
			log.Printf(color.FgGB("    %s: '%s' will change from '%v' to '%v'", r.Desc(), key, currentValue, value))
		}
	}

	return nil
}
