# service resource


## Actions

* `start`:

* `stop`:

* `restart`:

* `reload`:

* `enable`:

* `disable`:

## Attributes

* `name` (string or table) (default: name of resource):

* `provider` (string):


## Example

```lua
service "httpd" {
    action = {"start", "enable"},
}
```
