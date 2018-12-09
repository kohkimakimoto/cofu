# Cofu [![Build Status](https://travis-ci.org/kohkimakimoto/cofu.svg?branch=master)](https://travis-ci.org/kohkimakimoto/cofu)

Minimum configuration management tool written in Go.

## Installation

Cofu is provided as a single binary. You can download it and drop it in your $PATH.

[Download latest version](https://github.com/kohkimakimoto/cofu/releases/latest)

## Usage

```lua
$ echo '
software_package "nginx" {
    action = "install",
}

service "nginx" {
    action = {"enable", "start"},
}' > recipe.lua
$ sudo cofu recipe.lua
```

Usage is covered in more detail in the [Documentation](#documentation).

## Supported Platforms

* CentOS6
* CentOS7
* Debian8
* Debian9
* Ubuntu16.04
* Ubuntu18.04

## Documentation

If you're just getting started with Cofu, please start to read [Getting Started](docs/getting-started.md).

* [Getting Started](docs/getting-started.md)
* [Configuration](docs/configuration.md)
* [Resources](docs/resources.md)
    * [directory](docs/resources_directory.md)
    * [execute](docs/resources_execute.md)
    * [file](docs/resources_file.md)
    * [group](docs/resources_group.md)
    * [link](docs/resources_link.md)
    * [lua_function](docs/resources_lua_function.md)
    * [remote_directory](docs/resources_remote_directory.md)
    * [remote_file](docs/resources_remote_file.md)
    * [service](docs/resources_service.md)
    * [software_package](docs/resources_software_package.md)
    * [template](docs/resources_template.md)
    * [user](docs/resources_user.md)
    * [recipe](resources_recipe.md)
* [Variables](docs/variables.md)
* [Built-in Functions](docs/built-in-functions.md)
    * [include_recipe](docs/built-in-functions_include_recipe.md)
    * [run_command](docs/built-in-functions_run_command.md)
* [Built-in Libraries](docs/built-in-libraries.md)

## See Also

This software is highly inspired by [itamae-kitchen/itamae](https://github.com/itamae-kitchen/itamae).

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>

## License

The MIT License (MIT)
