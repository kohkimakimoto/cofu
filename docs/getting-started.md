# Getting Started

Cofu is a minimum configuration management tool written in Go.

This software is highly inspired by [itamae-kitchen/itamae](https://github.com/itamae-kitchen/itamae).

## Installation

Cofu is provided as a single binary. You can download it and drop it in your $PATH.

[Download latest version](https://github.com/kohkimakimoto/cofu/releases/latest)

After installing Cofu, run the `cofu` without any options in your terminal to check working.

```
$ cofu
Usage: cofu [OPTIONS...] [RECIPE_FILE] [ARGMENTS...]

  cofu -- Minimum configuration management tool.
  version 0.0.1 (3dbc091a9ab67a7572e9d5f97e901cef0816fa8d)

Options:
  -e 'command'               Execute 'command'
  -l, -log-level=LEVEL       Log level (quiet|error|warning|info|debug). Default is 'info'.
  -h, -help                  Show help
  -n, -dry-run               Runs dry-run mode
  -v, -version               Print the version
  -color                     Force ANSI output
  -no-color                  Disable ANSI output
  -var=JSON                  JSON string to input variables.
  -var-file=JSON_FILE        JSON file to input variables.

```

## Getting Started

Create a recipe file as `recipe.lua`, that is a configuration for assembling a server.

```lua
software_package "nginx" {
    action = "install",
}

service "nginx" {
    action = {"enable", "start"},
}
```
And then execute `cofu` command to apply a recipe to your machine.

```
$ sudo cofu recipe.lua
==> Starting cofu...
==> Loaded 2 resources.
==> Evaluating software_package[nginx]
    software_package[nginx]: installed will change from 'false' to 'true'
==> Evaluating service[nginx]
    service[nginx]: running will change from 'false' to 'true'
==> Complete!
```
