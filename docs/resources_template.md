# template resource

`template` is a resource to manage a file on file system with expanding [text/template](https://golang.org/pkg/text/template/).

## Actions

* `create`: (default).

* `delete`:

## Attributes

* `path` (string) (default: name of resource):

* `content` (string):

* `source` (string): Path to the file source. This is automatically configured by `path` attribute at default. For example, If the `path` is `/etc/php.ini`, `source` is `etc/php.ini` or `etc/php.ini.tmpl` that is a relative path from the directory which includes this recipe.

* `mode` (string):

* `owner` (string):

* `group` (string):

* `variables` (table):

## Example

```lua
template "/etc/php.ini" {
    mode = "0644",
    owner = "root",
    group = "root",
}
```

In a template, you can use [variables](variables.md):

```
{{var.hoge}}
```

And you can use `variable` attributes to pass the parameters.

```
template "/path/to/file" {
    variables = {
        foo = "Value of foo",
        bar = "Value of bar",
    },
}
```

templates:

```
{{foo}}
{{bar}}
```
