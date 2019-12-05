# Cofu Agent <!-- omit in toc -->

Cofu Agent is a flexible SSH server that is included in `cofu` binary.
It is especially useful for executing commands with a specific environment on a remote server. It will help you to provision remote server via SSH connection.

## Table of Contents <!-- omit in toc -->

- [Usage](#usage)
- [Functionalities](#functionalities)
  - [Sandboxes](#sandboxes)
  - [Environment Variables](#environment-variables)
  - [Functions](#functions)
- [Configuration](#configuration)
  - [Global section](#global-section)
    - [Global section parameters](#global-section-parameters)
    - [Global section example](#global-section-example)
  - [`functions.x` section](#functionsx-section)
    - [`functions.x` section parameters](#functionsx-section-parameters)
    - [`functions.x` section example](#functionsx-section-example)
  - [`include` section](#include-section)
    - [`include` section parameters](#include-section-parameters)
    - [`include` section example](#include-section-example)

## Usage

You need to create a configuration file to start Cofu Agent. Create the following config file:

```toml
authorized_keys = [
  # Replace this with your ssh public key.
  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBL...",
]
hot_reload = true
```

Then, run `cofu -agent` with the config:

```
$ cofu -agent -config-file=your-config.toml
```

If you get like the following messages, you are successful in starting Cofu Agent.

```
2019-12-04T16:29:02+09:00 cofu-agent INFO Starting cofu-agent
2019-12-04T16:29:02+09:00 cofu-agent INFO Sandbox directory: /tmp/cofu-agent/sandboxes
2019-12-04T16:29:02+09:00 cofu-agent INFO Using host key file /tmp/cofu-agent-host.key
2019-12-04T16:29:02+09:00 cofu-agent INFO Starting SSH protocol server on 0.0.0.0:2222
```

Try to connect with it by using SSH protocol.

```
$ ssh -p 2222 kohkimakimoto@localhost
```

## Functionalities

### Sandboxes

Cofu Agent is implemented for executing server management tasks such as provisioning. Therefore, it provides some addtional functionalities that a general SSH server doesn't have. Sandboxes are one of them.

A sandbox is a directory that is created when a SSH client connected to Cofu Agent. It is a unique directory of each SSH session. Since it is set as the current working directory of a SHELL, you can operate your tasks in a clean directory every time. Cofu Agent also remove old unused sandoxes automatically if you set it in the config.

### Environment Variables

It's often helpful to store some configuration values in environment variables. You can config arbitrary environment variables. Cofu Agent sets these to the ssh session processes.

### Functions

Functions define custom endpoint. Functions are executed when the SSH client connected to Cofu Agent. You can specify the commands, executing user and group in the function definitions. the function names are exposed as a ssh username.

## Configuration

The config file must be written in [TOML](https://github.com/toml-lang/toml). You can specify the config file by `-c` or `-config-file` option when Cofu Agent runs.

The following is an example:

```toml
log_level = "info"
addr = "0.0.0.0:2222"
authorized_keys_file = "/etc/cofu-agent/authorized_keys"
authorized_keys = []
host_key_file = "/etc/cofu-agent/host_key"
sandboxes_directory = "/var/lib/cofu-agent/sandboxes"
hot_reload = false
keep_sandboxes = 10
disable_local_auth = false
environment_file = "/etc/cofu-agent/environment"

[functions.php]
entrypoint = [ "/bin/bash", "-c", """
DOCKER_RUN_OPTIONS=""
if [[ -n "${COFU_AGENT_PTY}" ]]; then
  DOCKER_RUN_OPTIONS="-t"
fi
docker run --rm -i -v $PWD:/shared -w /shared ${DOCKER_RUN_OPTIONS} php:7.3 ${COFU_AGENT_FUNCTION_SESSION_COMMAND}
"""]
command = [ "/bin/bash" ]

[include]
files = [
  "/etc/cofu-agent/conf.d/*.toml"
]
```

### Global section

This section defines global settings for Cofu Agent server process.

#### Global section parameters

* `log_level` (string): The log level (`debug|info|warn|error`). The default is `info`. All logs in Cofu Agent outputs STDOUT.

* `addr` (string): The listen address to the Cofu Agent process. The default is `0.0.0.0:2222`.

* `authorized_keys_file` (string): Specifies the path of the authorized keys file to set authorization config.

* `authorized_keys` (string): Specifies the public keys directly in the config file to set authorization config.

* `disable_local_auth` (bool): If you set it `true`, Cofu Agent does not validate key when a client connects from the localhost. The default is `false`.

* `host_key_file` (string): Specifies the host key file path. If you set it and it does not exists, Cofu Agent generates new key file to the path and use it.

* `host_key` (string): Specifies the host key directly in the config file.

* `sandboxes_directory` (string): This is the parent directory of the sandbox directories.

* `keep_sandboxes` (int): Number of sandboxes for keeping. Cofu Agent removes old sandboxes automatically. If the value is `0`, does not remove any sandboxes.

* `environment` (array of string): Specifies the environment variables with such as `KEY=VALUE` format.

* `environment_file` (array of string): Specifies the file path. This file contains environment variables with such as `KEY=VALUE` format.

* `hot_reload` (bool): If you set it `true`, Cofu Agent reloads the config file every client requests.

#### Global section example

```toml
log_level = "info"
addr = "0.0.0.0:2222"
authorized_keys_file = "/etc/cofu-agent/authorized_keys"
authorized_keys = [
  "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBL...",
]
host_key_file = "/etc/cofu-agent/host_key"
sandboxes_directory = "/var/lib/cofu-agent/sandboxes"
keep_sandboxes = 10
disable_local_auth = false
environment = [
  "KEY1=VALUE1",
  "KEY2=VALUE2",
]
environment_file = "/etc/cofu-agent/environment"
hot_reload = false
```

### `functions.x` section

The `functions.x` section defines a function. `x` is the name of the function.

#### `functions.x` section parameters

* `authorized_keys_file` (string): Specifies the path of the authorized keys file to override authorization config for the function.

* `authorized_keys` (string): Specifies the public keys directly in the config file to override authorization config for the function.

* `user` (string): Specifies UNIX user to execute the function. The default is the user that runs Cofu Agent

* `group` (string): Specifies UNIX group to execute the function. The default is the group that runs Cofu Agent

* `entrypoint` (array of string): Specifies the executable and arguments such as `["executable", "param1", "param2"]`. When you connect the function Cofu Agent executes `entrypoint`.

* `command` (array of string): Specifies the default command that runs when you connect the function without any arguments such as `ssh -p 2222 example@localhost`. If you specify `entrypoint`. `command` is used to provide default arguments for the `entrypoint`.

* `max_processes` (int): Number of processes at the same time.

* `timeout` (int): The function is closed when the timeout elapses. The unit is second.

#### `functions.x` section example

```toml
[functions.php]
entrypoint = [ "/bin/bash", "-c", """
DOCKER_RUN_OPTIONS=""
if [[ -n "${COFU_AGENT_PTY}" ]]; then
  DOCKER_RUN_OPTIONS="-t"
fi
docker run --rm -i -v $PWD:/shared -w /shared ${DOCKER_RUN_OPTIONS} php:7.3 ${COFU_AGENT_FUNCTION_SESSION_COMMAND}
"""]
command = [ "/bin/bash" ]
```

Usage:

```
$ ssh -T -p 2222 php@localhost php -v
PHP 7.3.4 (cli) (built: Apr  6 2019 02:24:14) ( NTS )
Copyright (c) 1997-2018 The PHP Group
Zend Engine v3.3.4, Copyright (c) 1998-2018 Zend Technologies
```

You can also use an interactive bash shell in the docker container.

```
$ ssh -p 2222 php@localhost
```

### `include` section

The `include` section loads extra configuration files.

#### `include` section parameters

* `files` (array of strings): The files to be loaed within the configuration.

#### `include` section example

```toml
[include]
files = [
  "/etc/cofu-agent/conf.d/*.toml"
]
```
