# define

You can define ***definition*** with a collection of multiple resources, like:

```lua
define "install_and_enable_package" {
    version = "",
    function(params)
        software_package (params.name) {
            version = params.version,
        }

        service (params.name) {
            action = "enable",
        }
    end,
}

install_and_enable_package "httpd" {
    version = "2.4.6-40.el7.centos.1",
}
```

## What is a definition

A definition behaves like a compile-time macro that is reusable across recipes.
Though a definition looks like a resource, and at first glance seems like it could be used interchangeably, some important differences exist. A definition:

* Is not a resource or a custom resource
* Is processed while the resource collection is compiled.
* Does not support common resource properties, such as action, notifies, only_if, and not_if

A definition is like a macro. Therefore, before Cofu starts evaluating resources, all definitions are replaced to the resource collection by the definition's code.
