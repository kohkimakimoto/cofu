# Configuration

The configuration of Cofu are called **recipe**. The recipe describes the desired state of your server. It is mostly a collection of [resources](resources.md).
The recipe is written in [Lua](https://www.lua.org/) programming language.

## Example

```lua
software_package "nginx" {
    action = "install",
}

service "nginx" {
    action = {"enable", "start"},
}
```
