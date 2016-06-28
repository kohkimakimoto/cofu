# execute resource

`execute` is a resource to execute a command.

## Actions

* `run`: Run the command (default).

## Attributes

* `command` (string) (default: name of resource):

## Example

```lua
execute "create an empty file" {
  command = "touch /path/to/file",
  not_if = "test -e /path/to/file",
}
```
