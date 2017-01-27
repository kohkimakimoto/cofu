# remote_directory resource

`remote_directory` is a resource to transfer a directory from a source to a target.

## Actions

* `create`: (default).

* `delete`:

## Attributes

* `path` (string) (default: name of resource):

* `source` (string) (required):

* `mode` (string):

* `owner` (string):

* `group` (string):


## Example

```lua
remote_directory "/var/www/html" {
    source = "src/html",
}
```
