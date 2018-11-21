# Cofu [![Build Status](https://travis-ci.org/kohkimakimoto/cofu.svg?branch=master)](https://travis-ci.org/kohkimakimoto/cofu)

Minimum configuration management tool written in Go.

This software is highly inspired by [itamae-kitchen/itamae](https://github.com/itamae-kitchen/itamae).

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

Usage is covered in more detail in the [Documentation](./docs/README.md).

## Documentation

See [Documentation](./docs/README.md)

## Supported Platforms

* CentOS6
* CentOS7
* Ubuntu16.04
* Ubuntu18.04

## Developing Cofu

Requirements

* Go 1.11 or later (my development env)

Installing dependences

```
$ make deps
```

Building dev binary.

```
$ make dev
```

Building distributed binaries.


```
$ make dist
```

Building packages (now support only RPM)

```
$ make dist
$ make packaging
```

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>

## License

The MIT License (MIT)
