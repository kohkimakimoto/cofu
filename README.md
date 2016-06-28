# Cofu

Minimum configuration management tool written in Go.

This software is highly inspired by [itamae-kitchen/itamae](https://github.com/itamae-kitchen/itamae).

> **Now, it is on unstable stage. Sometimes API and code may be broken.**

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

## Notable Limitation!

I am testing Cofu on only el6 and el7 (CentOS6, CentOS7) in development. Therefore, Cofu does not officially support running on the other platforms.

But, to support running on multi platforms in future, Cofu is designed to execute different code by each platforms. This code is inspired by [mizzy/specinfra](https://github.com/mizzy/specinfra). If you want Cofu to support other platforms, try to port [mizzy/specinfra](https://github.com/mizzy/specinfra) code to Go code: https://github.com/kohkimakimoto/cofu/tree/master/infra

## Developing Cofu

Requirements

* Go 1.6 or later
* [Gom](https://github.com/mattn/gom)
* [direnv](https://github.com/direnv/direnv)

Installing dependences

```
$ make deps
```

Building dev binary.

```
$ make
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
