package resource

import (
	"github.com/kohkimakimoto/cofu/cofu"
	"github.com/yuin/gopher-lua"
)

var LuaFunction = &cofu.ResourceType{
	Name: "lua_function",
	Attributes: []cofu.Attribute{
		&cofu.StringSliceAttribute{
			Name:     "action",
			Default:  []string{"run"},
			Required: true,
		},
		&cofu.LFunctionAttribute{
			Name:     "func",
			Required: true,
		},
	},
	PreAction:                luaFunctionPreAction,
	SetCurrentAttributesFunc: luaFunctionSetCurrentAttributes,
	Actions: map[string]cofu.ResourceAction{
		"run": luaFunctionRunAction,
	},
}

func luaFunctionPreAction(r *cofu.Resource) error {
	r.Attributes["executed"] = true

	return nil
}

func luaFunctionSetCurrentAttributes(r *cofu.Resource) error {
	if r.CurrentAction == "run" {
		r.CurrentAttributes["executed"] = false
	}

	return nil
}

func luaFunctionRunAction(r *cofu.Resource) error {
	fn := r.GetLFunctionAttribute("func")

	L := r.App.LState
	err := L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    0,
		Protect: true,
	}, nil)
	if err != nil {
		return err
	}

	return nil
}
