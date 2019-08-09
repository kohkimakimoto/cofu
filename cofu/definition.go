package cofu

import (
	"github.com/yuin/gopher-lua"
)

type Definition struct {
	Name          string
	DefaultParams *lua.LTable
	Func          *lua.LFunction
	app           *App
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

	// copy attrs
	attrs.ForEach(func(k, v lua.LValue) {
		mergedAttrs.RawSet(k, v)
	})

	definition.DefaultParams.ForEach(func(k, v lua.LValue) {
		overwriteValue := mergedAttrs.RawGet(k)
		if overwriteValue == lua.LNil {
			// set default
			mergedAttrs.RawSet(k, v)
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
