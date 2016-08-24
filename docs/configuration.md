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

## Another Syntax

The above example of configuration is written in Lua DSL style.
You can also use plain Lua functions styles. The following examples are valid config code.

```lua
software_package("nginx", {
    action = "install",
})

service("nginx", {
    action = {"enable", "start"},
})
```

or

```lua
local package_nginx = software_package "nginx"
package_nginx.action = "install"

local nginx = service "nginx"
nginx.action = {"enable", "start"}
```
