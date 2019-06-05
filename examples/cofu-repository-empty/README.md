# example: cofu repository empty

This repository is used with cofu-agent and the following config:

```lua
[tasks.apply]
sandbox = true
sandbox_source = "github.com/kohkimakimoto/cofu/examples/cofu-repository-empty"
keep_sandboxes = 5
max_processes = 1
environment = [
  "HOME=/root",
  "USER=root",
]
command = [ "cofu", "-color", "./entrypoint.lua" ]
```
