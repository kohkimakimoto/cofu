# Built-in Libraries

Cofu includes below lua libraries.

* `glua.json`: Json encoder/decoder. It is implemented by [gluajson](https://github.com/kohkimakimoto/gluajson).
* `glua.fs`: Filesystem utility. It is implemented by [gluafs](https://github.com/kohkimakimoto/gluafs).
* `glua.yaml`: Yaml parser. It is implemented by [gluayaml](https://github.com/kohkimakimoto/gluayaml).
* `glua.template`: Text template. It is implemented by [gluatemplate](https://github.com/kohkimakimoto/gluatemplate).
* `glua.question`: A library to prompt the user for input. It is implemented by [gluaquestion](https://github.com/kohkimakimoto/gluaquestion).
* `glua.env`: Utility package for manipulating environment variables. It is implemented by [gluaenv](https://github.com/kohkimakimoto/gluaenv).
* `glua.http`: Http module. It is implemented by [gluahttp](https://github.com/cjoudrey/gluahttp).

## Example

```lua
local json = require "glua.json"
local jsonStr = json.encode({"a",1,"b",2,"c",3})

print(jsonstr)
```
