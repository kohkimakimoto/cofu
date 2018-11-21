# Built-in Libraries

Cofu includes below lua libraries.

* `json`: [layeh/gopher-json](https://github.com/layeh/gopher-json).
* `fs`: [kohkimakimoto/gluafs](https://github.com/kohkimakimoto/gluafs).
* `yaml`: [kohkimakimoto/gluayaml](https://github.com/kohkimakimoto/gluayaml).
* `question`: [kohkimakimoto/question](https://github.com/kohkimakimoto/gluaquestion).
* `template`: [kohkimakimoto/gluatemplate](https://github.com/kohkimakimoto/gluatemplate).
* `env`: [kohkimakimoto/gluaenv](https://github.com/kohkimakimoto/gluaenv).
* `http`: [cjoudrey/gluahttp](https://github.com/cjoudrey/gluahttp).
* `re`: [yuin/gluare](https://github.com/yuin/gluare)
* `sh`:[otm/gluash](https://github.com/otm/gluash)
* `crypto`:[tengattack/gluacrypto](https://github.com/tengattack/gluacrypto)

## Example

```lua
local json = require "json"
local jsonStr = json.encode({"a",1,"b",2,"c",3})

print(jsonstr)
```
