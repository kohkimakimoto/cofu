# resource resource

`resource` is a special function to load an another recipe as a resource. 
The difference between `resource` and `include_recipe` is that
`resource` evaluates a recipe file in the independent context, whereas
`include_recipe` evaluates a recipe file in the same context that loads it.

`resource` also loads a recipe under the `resources` directory if it exists.

## Actions

* `run`: Run the recipe (default).

## Attributes

* `path` (string) (default: name of resource):

And you can set arbitrary attributes to pass variables to a recipe.

## Example

```lua
resource "example" {
    variable1 = "aaa",
    variable2 = "bbb",
}
```

example.lua ( or `resources/example.lua`)

```lua
software_package "httpd"

template "/etc/httpd/conf/httpd.conf" {

}

-- you can use varilabes passed with `resource`. They are not global variables passed by `-var` or `-var-file` options.
print(var.variable1)
print(var.variable2) 
```
