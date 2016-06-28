# file resource

`file` is a resource to manage a file on file system.

## Actions

* `create`: (default).

* `delete`:

## Attributes

* `path` (string) (default: name of resource):

* `content` (string):

* `mode` (string):

* `owner` (string):

* `group` (string):

## Example

```lua
file "/path/to/file" {
    content = [=[
Hello world!


]=],
    mode = "0644",
    owner = "root",
    group = "root",
}
```
