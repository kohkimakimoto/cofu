# recipe resource

`recipe` loads an another recipe as resource. 
The difference between `recipe` and `include_recipe` is that
`recipe` evaluate a recipe file in the independent context, whereas
`include_recipe` evaluate a recipe file in the same context that loads it.

## Actions

* `run`: Run the recipe (default).

## Attributes

* `path` (string) (default: name of resource):
* `variables` (table):

## Example

```lua
recipe "cron" {
  variables = {

  }
}
```
