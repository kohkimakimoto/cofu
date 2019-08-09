package cofu

import (
	"encoding/json"
	"errors"
	"fmt"
	fatihColor "github.com/fatih/color"
	"github.com/kohkimakimoto/cofu/infra"
	"github.com/kohkimakimoto/cofu/support/color"
	"github.com/kohkimakimoto/loglv"
	"github.com/labstack/gommon/log"
	"github.com/yookoala/realpath"
	"github.com/yuin/gopher-lua"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type App struct {
	LState               *lua.LState
	Logger               Logger
	ResourceTypes        []*ResourceType
	ResourceTypesMap     map[string]*ResourceType
	Resources            []*Resource
	DelayedNotifications []*Notification
	Infra                *infra.Infra
	DryRun               bool
	Tmpdir               string
	Tmpfiles             []string
	variable             map[string]interface{}
	Parent               *App
	Level                int
	LogHeader            string
	BuiltinRecipes       map[string]string
	Basepath             string
}

const LUA_APP_KEY = "*__COFU_APP__"

func NewApp() *App {
	defaultLogHeader := `${level} ${prefix}`
	defaultLogger := log.New("cofu")
	defaultLogger.SetPrefix("")
	defaultLogger.SetHeader(defaultLogHeader)

	return &App{
		LState:               lua.NewState(),
		Logger:               defaultLogger,
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
		Parent:         nil,
		Level:          0,
		LogHeader:      defaultLogHeader,
		Basepath:       "",
	}
}

func (app *App) Init() error {
	// load resource types and define lua functions.
	for _, resourceType := range app.ResourceTypes {
		if err := app.loadResourceType(resourceType); err != nil {
			return err
		}
	}

	// load lua libraries.
	openLibs(app)

	// register app into the LState
	ud := app.LState.NewUserData()
	ud.Value = app

	app.LState.SetGlobal(LUA_APP_KEY, ud)
	app.LState.SetGlobal("var", toLValue(app.LState, app.variable))

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

	if app.Parent != nil {
		app.Logger.SetPrefix(GenLogIndent(app.Parent.Level))
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

func (app *App) LoadVariableFromMap(m map[string]interface{}) error {
	variable := app.variable
	for k, v := range m {
		variable[k] = v
	}
	app.variable = variable

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
	return loadRecipeFile(recipeFile, app.LState, app)
}

func (app *App) RemoveDuplicateDelayedNotification() {
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

func (app *App) Run(dryRun bool) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	logger := app.Logger

	app.DryRun = dryRun

	// create tmp directory
	if _, err := os.Stat(app.Tmpdir); os.IsNotExist(err) {
		defaultUmask := syscall.Umask(0)
		os.MkdirAll(app.Tmpdir, 0777)
		syscall.Umask(defaultUmask)
	}

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

	if app.IsRootApp() {
		logger.Info("Starting " + Name + "...")
	}

	if app.DryRun && app.IsRootApp() {
		logger.Info(color.FgCB("Running on dry-run mode. It does not affect any real resources."))
	}

	if app.IsRootApp() {
		logger.Debugf("Log level '%s'", loglv.LvString())
		logger.Debugf("os_family '%s'", app.Infra.Command().OSFamily())
		logger.Debugf("os_release '%s'", app.Infra.Command().OSRelease())
	}

	logger.Debugf("Loaded %d resource(s).", len(app.Resources))

	for _, r := range app.Resources {
		err := r.Run("")
		if err != nil {
			return err
		}
	}

	app.RemoveDuplicateDelayedNotification()
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

	if app.IsRootApp() {
		logger.Info("Complete!")
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

func (app *App) IsRootApp() bool {
	return app.Level == 0
}

func (app *App) LoadUserResources() error {
	return nil
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

func GetLogger(L *lua.LState) (Logger, error) {
	app, err := GetApp(L)
	if err != nil {
		return nil, err
	}

	return app.Logger, nil
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

func GenLogIndent(level int) string {
	return fmt.Sprintf("%s", strings.Repeat("  ", level))
}

func SetNoColor(b bool) {
	fatihColor.NoColor = false
}

func getBasepath(L *lua.LState) string {
	basepath, err := basepath(L)
	if err != nil {
		wd, err2 := os.Getwd()
		if err2 == nil {
			basepath = wd
		}
	}

	return basepath
}

func loadRecipeFile(recipeFile string, L *lua.LState, app *App) error {
	orgBase := app.Basepath
	defer func() {
		app.Basepath = orgBase
	}()

	dir := filepath.Dir(recipeFile)
	base, err := realpath.Realpath(dir)
	if err != nil {
		return err
	}
	app.Basepath = base

	if err := L.DoFile(recipeFile); err != nil {
		return err
	}

	return nil
}
