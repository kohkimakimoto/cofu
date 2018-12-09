# Resources

A resource is a statement of configuration policy that describes the desired state for an item.

## Syntax

```lua
resource_type "name" {
   attribute = "value",
   action = "type_of_action",
}
```

For Example: a resource that is used to install a nginx package may look something like this:

```lua
software_package "nginx" {
    action = "install",
}
```

## Resource Type

* [directory](resources_directory.md)
* [execute](resources_execute.md)
* [file](resources_file.md)
* [group](resources_group.md)
* [link](resources_link.md)
* [lua_function](resources_lua_function.md)
* [remote_file](resources_remote_directory.md)
* [remote_file](resources_remote_file.md)
* [service](resources_service.md)
* [software_package](resources_software_package.md)
* [template](resources_template.md)
* [user](resources_user.md)
* [recipe](resources_recipe.md)

## Common Attributes

All resource types have the following common attributes.

* `only_if` (string): If only_if command exits with non-zero status, the resource will not be executed.

* `not_if` (string): If not_if command exits with zero status, the resource will not be executed.

* `user` (string): If you specified this, commands related with the resource will be executed as the user.

* `cwd` (string): If you specified this, commands related with the resource will be executed on the working directory.

* `notifies`: (table): If you specified this, Cofu runs other resources when the resource is updated. The syntax is like the following.

  restart httpd service:

  ```lua
  notifies = {"restart", "service[httpd]"}
  ```

  restart httpd service immediately:

  ```lua
  notifies = {"restart", "service[httpd]", "immediately"}
  ```

  restart httpd, nginx service. httpd service restarted immediately:

  ```lua
  notifies = {{"restart", "service[httpd]", "immediately"}, {"restart", "service[nginx]"}}
  ```
  
* `verify` (string or table): If you specified this, runs the commands. If the result of the commands is non-zero status, Cofu exits with error.

* `description` (string): If you specified this, Cofu show this description when the resource evaluates. 

## Common Actions

All resource types support the following common actions.

* `nothing`: Nothing to do.
