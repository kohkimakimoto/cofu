# remote_file resource

`remote_file` is a resource to transfer a file from a source to a target.

## Actions

* `create`: (default).

* `delete`:

## Attributes

* `path` (string) (default: name of resource):

* `source` (string): Path to the file source. This is automatically configured by `path` attribute at default. For example, If the `path` is `/etc/php.ini`, `source` is `etc/php.ini` that is a relative path from the directory which includes this recipe.

* `mode` (string):

* `owner` (string):

* `group` (string):


## Example

```lua
remote_file "/etc/php.ini" {

}
```
