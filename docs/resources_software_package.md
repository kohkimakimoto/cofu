# software_package resource


## Actions

* `install`: (default).

* `remove`:

## Attributes

* `name` (string or table) (default: name of resource):

* `version` (string):

* `options` (string):


## Example

```lua
software_package "httpd" {
    action = "install",
}
```
