# lua_function resource

`lua_function` is a resource to execute lua function.

## Actions

* `run`: (default).

## Attributes

* `func` (function):

## Example

```lua
lua_function "run_luafunc" {
    func = function()

        print("hogehoge")

    end,
}
```
