# directory resource

`directory` is a resource to manage a directory on file system.

## Actions

* `create`: Create the directory (default).

* `delete`: Delete the directory.

## Attributes

* `path` (string) (default: name of resource):

* `mode` (string):

* `owner` (string):

* `group` (string):

## Example

```lua
directory "/path/to/directory" {
    action = "create",
    mode = "0755",
    owner = "root",
    group = "root",
}
```
