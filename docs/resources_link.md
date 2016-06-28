# link resource

`link` is a resource to ensure a symlink exists with specified source and destination.

## Actions

* `create`: (default).

## Attributes

* `link` (string) (default: name of resource):

* `to` (string):

* `force` (bool):

## Example

```lua
link "/path/to/link" {
    to = "/path/to/source",
}
```
