package cofu

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/kohkimakimoto/cofu/infra"
	"github.com/kohkimakimoto/cofu/support/color"
	"github.com/kohkimakimoto/loglv"
	"github.com/yuin/gopher-lua"
)

type App struct {
	LState               *lua.LState
	LogLevel             string
	DryRun               bool
	ResourceTypes        []*ResourceType
	ResourceTypesMap     map[string]*ResourceType
	Resources            []*Resource
	DelayedNotifications []*Notification
	Infra                *infra.Infra
	Tmpdir               string
	Tmpfiles             []string
	variable             map[string]interface{}
}

const LUA_APP_KEY = "*__COFU_APP__"

func NewApp() *App {
	L := lua.NewState()
	app := &App{
		LState:               L,
		ResourceTypesMap:     map[string]*ResourceType{},
		Resources:            []*Resource{},
		DelayedNotifications: []*Notification{},
		Infra:                infra.New(),
		Tmpdir:               "/tmp/cofu_tmp",
		Tmpfiles:             []string{},
		variable: map[string]interface{}{
			"GOARCH": runtime.GOARCH,
			"GOOS":   runtime.GOOS,
		},
	}

	ud := L.NewUserData()
	ud.Value = app

	L.SetGlobal(LUA_APP_KEY, ud)
	L.SetGlobal("var", toLValue(L, app.variable))

	return app
}

func (app *App) Init() error {
	// It intends not to output timestamp with log.
	log.SetFlags(0)
	// support leveled logging.
	loglv.Init()
	// output to stdout
	loglv.SetOutput(os.Stdout)

	if app.LogLevel == "" {
		app.LogLevel = "info"
	}

	err := loglv.SetLevelByString(app.LogLevel)
	if err != nil {
		return err
	}

	// load resource types and define lua functions.
	for _, resourceType := range app.ResourceTypes {
		if err := app.loadResourceType(resourceType); err != nil {
			return err
		}
	}

	// load lua libraries.
	openLibs(app)

	// create tmp directory
	if _, err := os.Stat(app.Tmpdir); os.IsNotExist(err) {
		defaultUmask := syscall.Umask(0)
		os.MkdirAll(app.Tmpdir, 0777)
		syscall.Umask(defaultUmask)
	}

	return nil
}

func (app *App) loadResourceType(resourceType *ResourceType) error {
	if _, ok := app.ResourceTypesMap[resourceType.Name]; ok {
		return fmt.Errorf("Already defined resource type '%s'", resourceType.Name)
	}
	app.ResourceTypesMap[resourceType.Name] = resourceType

	// set app reference.
	resourceType.app = app

	// append common attributes
	resourceType.Attributes = append(resourceType.Attributes, CommonAttributes...)

	// append common action
	if resourceType.Actions == nil {
		resourceType.Actions = map[string]ResourceAction{}
	}
	resourceType.Actions["nothing"] = func(r *Resource) error {
		return nil
	}

	return nil
}

func (app *App) LoadDefinition(definition *Definition) {
	// set lua api
	L := app.LState
	L.SetGlobal(definition.Name, L.NewFunction(definition.LGFunction()))
}

func (app *App) Close() {
	app.LState.Close()
	for _, f := range app.Tmpfiles {
		os.RemoveAll(f)
	}
}

func (app *App) LoadVariableFromJSON(v string) error {
	variable := app.variable
	err := json.Unmarshal([]byte(v), &variable)
	if err != nil {
		return err
	}

	L := app.LState
	L.SetGlobal("var", toLValue(L, app.variable))

	return nil
}

func (app *App) LoadVariableFromJSONFile(jsonFile string) error {
	b, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return err
	}

	variable := app.variable
	err = json.Unmarshal(b, &variable)
	if err != nil {
		return err
	}

	L := app.LState
	L.SetGlobal("var", toLValue(L, app.variable))

	return nil
}

func (app *App) LoadRecipe(recipeContent string) error {
	if err := app.LState.DoString(recipeContent); err != nil {
		return err
	}

	return nil
}

func (app *App) LoadRecipeFile(recipeFile string) error {
	if err := app.LState.DoFile(recipeFile); err != nil {
		return err
	}

	return nil
}

func (app *App) RemoveDuplicateDelayeNotification() {
	newDelayedNotifications := []*Notification{}

	for _, n := range app.DelayedNotifications {
		dup := false

		for _, nn := range newDelayedNotifications {
			if nn.TargetResourceDesc == n.TargetResourceDesc && nn.Action == n.Action {
				// duplication
				dup = true
			}
		}

		if !dup {
			newDelayedNotifications = append(newDelayedNotifications, n)
		}
	}

	app.DelayedNotifications = newDelayedNotifications
}

func (app *App) EnqueueDelayedNotification(n *Notification) {
	app.DelayedNotifications = append(app.DelayedNotifications, n)
}

func (app *App) DequeueDelayedNotification() *Notification {
	if len(app.DelayedNotifications) == 0 {
		return nil
	}

	// shift slice: https://github.com/golang/go/wiki/SliceTricks
	a := app.DelayedNotifications
	n, a := a[0], a[1:]
	app.DelayedNotifications = a

	return n
}

func (app *App) Run() error {
	if len(app.Resources) == 0 {
		// not found available resources.
		return nil
	}

	// preprocess for resources
	for _, r := range app.Resources {
		// checks required attributes
		for _, definedAttribute := range r.ResourceType.Attributes {
			if definedAttribute.IsRequired() {
				if _, ok := r.Attributes[definedAttribute.GetName()]; !ok {
					return fmt.Errorf("resource '%s': '%s' attribute is required but it is not set.", r.Desc(), definedAttribute.GetName())
				}
			}
		}

		// parse and validate notifies attribute
		if r.GetRawAttribute("notifies") != nil {
			r.Notifications = r.GetRawAttribute("notifies").([]*Notification)
			for _, n := range r.Notifications {
				n.DefinedInResource = r
				if err := n.Validate(); err != nil {
					return err
				}
			}
		}

		// set default diff function it it does not have specific func.
		if r.ResourceType.ShowDifferences == nil {
			r.ResourceType.ShowDifferences = DefaultShowDifferences
		}
	}

	if loglv.IsInfo() {
		log.Print("==> Starting " + Name + "...")
	}

	if loglv.IsDebug() {
		log.Printf("    (Debug) Log level '%s'", loglv.LvString())
		log.Printf("    (Debug) os_family '%s'", app.Infra.Command().OSFamily())
		log.Printf("    (Debug) os_release '%s'", app.Infra.Command().OSRelease())
	}

	if loglv.IsInfo() {
		log.Printf("==> Loaded %d resources.", len(app.Resources))
	}

	if app.DryRun {
		if loglv.IsInfo() {
			log.Print(color.FgCB("    Running on dry-run mode. It does not affect any real resources."))
		}
	}

	for _, r := range app.Resources {
		err := r.Run("")
		if err != nil {
			return err
		}
	}

	app.RemoveDuplicateDelayeNotification()
	for {
		n := app.DequeueDelayedNotification()
		if n == nil {
			break
		}

		err := n.Run()
		if err != nil {
			return err
		}
	}

	if loglv.IsInfo() {
		log.Printf("==> Complete!")
	}

	return nil
}

func (app *App) RegisterResource(r *Resource) {
	app.Resources = append(app.Resources, r)
}

func (app *App) FindOneResource(desc string) *Resource {
	for _, r := range app.Resources {
		if r.Desc() == desc {
			return r
		}
	}

	return nil
}

func (app *App) SendContentToTempfile(content []byte) (string, error) {
	tmpFile, err := ioutil.TempFile(app.Tmpdir, "")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.Write(content)
	if err != nil {
		return "", err
	}

	err = tmpFile.Chmod(0600)
	if err != nil {
		return "", err
	}

	app.Tmpfiles = append(app.Tmpfiles, tmpFile.Name())

	return tmpFile.Name(), nil
}

func (app *App) SendFileToTempfile(file string) (string, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return app.SendContentToTempfile(content)
}

func (app *App) SendDirectoryToTempDirectory(src string) (string, error) {
	tmpDir, err := ioutil.TempDir(app.Tmpdir, "")
	if err != nil {
		return "", err
	}

	tmpDir2 := filepath.Join(tmpDir, fmt.Sprintf("%d", time.Now().Unix()))

	err = CopyDir(src, tmpDir2)
	if err != nil {
		return "", err
	}

	app.Tmpfiles = append(app.Tmpfiles, tmpDir)
	return tmpDir2, nil
}

func GetApp(L *lua.LState) (*App, error) {
	ud, ok := L.GetGlobal(LUA_APP_KEY).(*lua.LUserData)
	if !ok {
		return nil, errors.New("Couldn't get a app object from LState. Your global variable '" + LUA_APP_KEY + "' was broken!")
	}

	app, ok := ud.Value.(*App)
	if !ok {
		return nil, errors.New("Your global variable '" + LUA_APP_KEY + "' was broken!")
	}

	return app, nil
}

func toLValue(L *lua.LState, value interface{}) lua.LValue {
	switch converted := value.(type) {
	case bool:
		return lua.LBool(converted)
	case float64:
		return lua.LNumber(converted)
	case string:
		return lua.LString(converted)
	case []interface{}:
		arr := L.CreateTable(len(converted), 0)
		for _, item := range converted {
			arr.Append(toLValue(L, item))
		}
		return arr
	case map[string]interface{}:
		tbl := L.CreateTable(0, len(converted))
		for key, item := range converted {
			tbl.RawSetH(lua.LString(key), toLValue(L, item))
		}
		return tbl
	}
	return lua.LNil
}
