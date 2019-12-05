# Cofu Agent

Cofu Agent is a flexible SSH server that is included in `cofu` binary.
It is especially useful for executing commands with a specific environment on a remote server. It will help you to provision remote server via SSH connection.

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
$ ssh -T -p 2222 kohkimakimoto@localhost
```

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
