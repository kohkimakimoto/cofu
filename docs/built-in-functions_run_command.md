# run_command

Run commands in a recipe.

## Example

```lua
local result = run_command("echo -n Hello")
print(result:stdout())
print(result:stderr())
print(result:combined())
print(result:exit_status())
print(result:failure())
print(result:success())
```
