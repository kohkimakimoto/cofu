package cofu

import (
	"github.com/yuin/gopher-lua"
)

// Definition is deprecated from v0.8.0

type Definition struct {
	Name   string
	Params *lua.LTable
	Func   *lua.LFunction
	app    *App
}

func (definition *Definition) LGFunction() func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		// procedural style
		if L.GetTop() == 2 {
			tb := L.CheckTable(2)
			definition.run(L, name, tb)

			return 0
		}

		// DSL style
		L.Push(L.NewFunction(func(L *lua.LState) int {
			tb := L.CheckTable(1)
			definition.run(L, name, tb)

			return 0
		}))

		return 1
	}
}

func (definition *Definition) run(L *lua.LState, name string, attrs *lua.LTable) {
	mergedAttrs := L.NewTable()

	definition.Params.ForEach(func(k, v lua.LValue) {
		overwriteValue := attrs.RawGet(k)
		if overwriteValue == lua.LNil {
			mergedAttrs.RawSet(k, v)
		} else {
			mergedAttrs.RawSet(k, overwriteValue)
		}
	})

	mergedAttrs.RawSetString("name", lua.LString(name))

	err := L.CallByParam(lua.P{
		Fn:      definition.Func,
		NRet:    0,
		Protect: true,
	}, mergedAttrs)
	if err != nil {
		panic(err)
	}
}
