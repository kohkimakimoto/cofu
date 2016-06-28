# Variables

You can input variables to a recipe by `-var` and `-var-file` option.
The variables can be read in recipes by `var` global variable.

## Example

Create `recipe.lua`.

```lua
user (var.name) {
    uid = 1020,
}
```

Run cofu with `-var`.

```
$ cofu recipe.lua -var='{"name": "kohkimakimoto"}'
```

or. create JSON file. and load it as the following.

```
$ cofu recipe.lua -var-file=var.json
```
