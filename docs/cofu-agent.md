# Cofu Agent <!-- omit in toc -->

Cofu Agent is a flexible SSH server that is included in `cofu` binary.
It is especially useful for executing commands with a specific environment on a remote server. It will help you to provision remote server via SSH connection.

## Table of Contents <!-- omit in toc -->

- [Usage](#usage)
- [Sandboxes](#sandboxes)
- [Functions](#functions)
- [Configuration](#configuration)
  - [Global section](#global-section)
    - [Global section parameters](#global-section-parameters)
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

## Sandboxes

Cofu Agent is implemented for executing server management tasks such as provisioning. Therefore, it provides some addtional functionalities that a general SSH server doesn't have. Sandboxes are one of them.

A sandbox is a directory that is created when a SSH client connected to Cofu Agent. The sandbox is a unique directory of each SSH session. It is set as the current working directory of a SHELL.

## Functions

WIP...

## Configuration

The config file must be written in [TOML](https://github.com/toml-lang/toml). You can specify the config file by `-c` or `-config-file` option when Cofu Agent runs.

The following is an example:

```toml
# log_level
log_level = "info"

# addr
addr = "0.0.0.0:2222"

# authorized_keys_file
authorized_keys_file = "/etc/cofu-agent/authorized_keys"

# authorized_keys
authorized_keys = []

# host_key_file
host_key_file = "/etc/cofu-agent/host_key"

# sandboxes_directory
sandboxes_directory = "/var/lib/cofu-agent/sandboxes"

# hot_reload
hot_reload = false

# keep_sandboxes
keep_sandboxes = 10

# disable_local_auth. This config is for development purpose. You should not set true in the production environment.
disable_local_auth = false

# environment_file
environment_file = "/etc/cofu-agent/environment"

# environment
# environment = []
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

### `functions.x` section

The `functions.x` section defines a function. `x` is the name of the function.

#### `functions.x` section parameters

WIP...

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
