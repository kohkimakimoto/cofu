# resource resource

`resource` loads an another recipe as a resource. 
The difference between `resource` and `include_recipe` is that
`resource` evaluates a recipe file in the independent context, whereas
`include_recipe` evaluates a recipe file in the same context that loads it.

## Actions

* `run`: Run the recipe (default).

## Attributes

* `path` (string) (default: name of resource):

And you can set arbitrary attributes to path variables recipe.

## Example

```lua
recipe "example" {
    variable1 = "aaa",
    variable2 = "bbb",
}
```

example.lua

```lua
execute "ls -la"


-- you can use varilabes
print(var.variable1)
print(var.variable2) 
```
