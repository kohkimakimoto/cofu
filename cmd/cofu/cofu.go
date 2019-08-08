package main

import (
	"flag"
	"fmt"
	"github.com/kohkimakimoto/cofu/_work/tmp"
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/kohkimakimoto/cofu/ext/agent"
	"github.com/kohkimakimoto/cofu/fetcher"
	"github.com/kohkimakimoto/cofu/resource"
	"github.com/kohkimakimoto/cofu/support/color"
	"github.com/kohkimakimoto/cofu/support/logutil"
	"github.com/labstack/gommon/log"
	"os"
)

func main() {
	os.Exit(realMain())
}

func realMain() (status int) {
	defer func() {
		if err := recover(); err != nil {
			//fmt.Fprint(os.Stderr, color.FgRB("runtime panic\n"))
			printError(err)
			status = 1
		}
	}()

	// parse flags...
	var optE, optLogLevel, optVarJson, optVarJsonFile, optConfigFile string
	var optVersion, optDryRun, optColor, optNoColor, optAgent, optFetch bool

	flag.StringVar(&optE, "e", "", "")
	flag.StringVar(&optLogLevel, "l", "info", "")
	flag.StringVar(&optLogLevel, "log-level", "info", "")
	flag.StringVar(&optVarJson, "var", "", "")
	flag.StringVar(&optVarJsonFile, "var-file", "", "")

	flag.BoolVar(&optDryRun, "n", false, "")
	flag.BoolVar(&optDryRun, "dry-run", false, "")
	flag.BoolVar(&optVersion, "v", false, "")
	flag.BoolVar(&optVersion, "version", false, "")

	flag.BoolVar(&optColor, "color", false, "")
	flag.BoolVar(&optNoColor, "no-color", false, "")

	// server|agent options
	//flag.BoolVar(&optServer, "s", false, "")
	//flag.BoolVar(&optServer, "server", false, "")
	flag.BoolVar(&optAgent, "a", false, "")
	flag.BoolVar(&optAgent, "agent", false, "")
	flag.StringVar(&optConfigFile, "c", "", "")
	flag.StringVar(&optConfigFile, "config-file", "", "")

	// hidden flag. run a sandbox fetcher
	flag.BoolVar(&optFetch, "fetch", false, "")

	flag.Usage = func() {
		fmt.Println(`Usage: ` + cofu.Name + ` [OPTIONS...] [RECIPE_FILE]

` + cofu.Name + ` -- ` + cofu.Usage + `
version ` + cofu.Version + ` (` + cofu.CommitHash + `)

Options:
  -e 'recipe'                Execute 'recipe'
  -l, -log-level=LEVEL       Log level (error|warn|info|debug). Default is 'info'.
  -h, -help                  Show help
  -n, -dry-run               Runs dry-run mode
  -v, -version               Print the version
  -color                     Force ANSI output
  -no-color                  Disable ANSI output
  -a, -agent                 Runs cofu agent
  -c, -config-file=FILE      Load agent config from the FILE
  -var=JSON                  JSON string to input variables.
  -var-file=JSON_FILE        JSON file to input variables.
`)
	}
	flag.Parse()

	if optVersion {
		// show version
		fmt.Println(cofu.Name + " version " + cofu.Version + " (" + cofu.CommitHash + ")")
		return 0
	}

	if optFetch {
		if err := doFetch(); err != nil {
			printError(err)
			return 1
		}
		return 0
	}

	//if optServer {
	//	// run server
	//	if err := server.Start(optConfigFile); err != nil {
	//		printError(err)
	//		return 1
	//	}
	//
	//	return 0
	//}

	if optAgent {
		// run agent
		if err := agent.Start(optConfigFile); err != nil {
			printError(err)
			return 1
		}

		return 0
	}

	if optE == "" && flag.NArg() == 0 {
		// show usage
		flag.Usage()
		return 0
	}

	var recipeFile string
	var recipeContent string

	if optE != "" {
		recipeContent = optE
	}

	if optE == "" && flag.NArg() > 0 {
		// specify the recipe file. parse flags again for using flags after the recipe file.
		recipeFile = flag.Arg(0)
		indexOfScript := (len(os.Args) - flag.NArg())
		flag.CommandLine.Parse(os.Args[indexOfScript+1:])
	}

	// setup the cofu app.
	app := cofu.NewApp()
	defer app.Close()

	// setup logger
	lv, err := logutil.LoglvlFromString(optLogLevel)
	if err != nil {
		printError(err)
		status = 1
	}

	logger := log.New("cofu")
	logger.SetLevel(lv)
	logger.SetPrefix("")
	logger.SetHeader(`${level}${prefix}`)
	if optColor {
		cofu.SetNoColor(false)
		logger.EnableColor()
	}
	if optNoColor {
		cofu.SetNoColor(true)
		logger.DisableColor()
	}
	app.Logger = logger

	app.ResourceTypes = resource.ResourceTypes
	app.BuiltinRecipes = tmp.DefaultBuiltinRecipes

	if optVarJsonFile != "" {
		if err := app.LoadVariableFromJSONFile(optVarJsonFile); err != nil {
			printError(err)
			return 1
		}
	}

	if optVarJson != "" {
		if err := app.LoadVariableFromJSON(optVarJson); err != nil {
			printError(err)
			return 1
		}
	}

	// initialize app
	if err := app.Init(); err != nil {
		printError(err)
		status = 1
	}

	if recipeFile != "" {
		if err := app.LoadRecipeFile(recipeFile); err != nil {
			printError(err)
			return 1
		}
	} else if recipeContent != "" {
		if err := app.LoadRecipe(recipeContent); err != nil {
			printError(err)
			return 1
		}
	}

	// run converging phase.
	if err := app.Run(optDryRun); err != nil {
		printError(err)
		return 1
	}

	return status
}

func doFetch() error {
	if len(os.Args) != 4 {
		return fmt.Errorf("usage: cofu -fetch [src] [dst]")
	}

	src := os.Args[2]
	dst := os.Args[3]

	fet := fetcher.NewFetcher()
	if err := fet.Fetch(src, dst); err != nil {
		return err
	}

	return nil
}

func printError(err interface{}) {
	fmt.Fprintf(os.Stderr, color.FgRB(cofu.Name+" aborted! "))
	fmt.Fprintf(os.Stderr, color.FgRB("%v\n", err))
}
