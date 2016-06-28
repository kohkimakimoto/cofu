# group resource

`group` is a resource to ensure a group exists.

## Actions

* `create`: (default).

## Attributes

* `groupname` (string) (default: name of resource):

* `gid` (number):

## Example

```lua
group "mygroup" {
    gid = 1010,
}
```
