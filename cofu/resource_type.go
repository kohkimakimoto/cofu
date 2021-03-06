package cofu

import (
	"fmt"
	"github.com/yuin/gopher-lua"
)

type ResourceType struct {
	Name                     string
	Attributes               []Attribute
	PreAction                ResourceAction
	SetCurrentAttributesFunc ResourceAction
	Actions                  map[string]ResourceAction
	ShowDifferences          ResourceAction
	UseFallbackAttributes    bool
	app                      *App
}

func (resourceType *ResourceType) LGFunction() func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		if L.GetTop() == 1 {
			// object or DSL style
			r := resourceType.registerResource(L, name)
			L.Push(newLResource(L, r))

			return 1
		} else if L.GetTop() == 2 {
			// function style
			tb := L.CheckTable(2)
			r := resourceType.registerResource(L, name)
			setupResource(r, tb)
			L.Push(newLResource(L, r))

			return 1
		}

		return 1
	}
}

func (resourceType *ResourceType) registerResource(L *lua.LState, name string) *Resource {
	app, err := GetApp(L)
	if err != nil {
		L.RaiseError(err.Error())
		return nil
	}

	r := NewResource(name, resourceType, app)

	// set default attributes
	for _, definedAttribute := range resourceType.Attributes {
		if definedAttribute.HasDefault() {
			r.Attributes[definedAttribute.GetName()] = definedAttribute.GetDefault()
		}

		switch a := definedAttribute.(type) {
		case *StringAttribute:
			if a.IsDefaultName() {
				r.Attributes[definedAttribute.GetName()] = r.Name
			}
		case *StringSliceAttribute:
			if a.IsDefaultName() {
				r.Attributes[definedAttribute.GetName()] = r.Name
			}
		}
	}

	app.RegisterResource(r)

	return r
}

// Lua Resource Class
const lResourceClass = "Resource*"

func loadLResourceClass(L *lua.LState) {
	mt := L.NewTypeMetatable(lResourceClass)

	L.SetField(mt, "__call", L.NewFunction(resourceCall))
	L.SetField(mt, "__index", L.NewFunction(resourceIndex))
	L.SetField(mt, "__newindex", L.NewFunction(resourceNewindex))
}

func updateResource(r *Resource, attributeName string, value lua.LValue) {
	var attribute Attribute
	for _, definedAttribute := range r.ResourceType.Attributes {
		if definedAttribute.GetName() == attributeName {
			attribute = definedAttribute
			break
		}
	}

	var goValue interface{}
	if attribute == nil {
		if r.ResourceType.UseFallbackAttributes {
			goValue = toGoValue(value)
			r.FallbackAttributes[attributeName] = goValue
		} else {
			panic(fmt.Sprintf("Invalid attribute name '%s'.", attributeName))
		}
	} else {
		goValue = attribute.ToGoValue(value)
	}

	r.AttributesLValues[attributeName] = value
	r.Attributes[attributeName] = goValue
}

func setupResource(r *Resource, attributes *lua.LTable) {
	attributes.ForEach(func(k, v lua.LValue) {
		if kstr, ok := toString(k); ok {
			updateResource(r, kstr, v)
		} else {
			panic(fmt.Sprintf("'%s' An attribute must be string", r.Desc()))
		}
	})
}

func newLResource(L *lua.LState, r *Resource) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = r
	L.SetMetatable(ud, L.GetTypeMetatable(lResourceClass))

	return ud
}

func checkResource(L *lua.LState) *Resource {
	ud := L.CheckUserData(1)
	if result, ok := ud.Value.(*Resource); ok {
		return result
	}
	L.ArgError(1, "Resource expected")

	return nil
}

func resourceCall(L *lua.LState) int {
	r := checkResource(L)
	tb := L.CheckTable(2)

	setupResource(r, tb)

	return 0
}

func resourceIndex(L *lua.LState) int {
	r := checkResource(L)
	index := L.CheckString(2)

	v, ok := r.AttributesLValues[index]
	if v == nil || !ok {
		v = lua.LNil
	}

	L.Push(v)
	return 1
}

func resourceNewindex(L *lua.LState) int {
	r := checkResource(L)
	index := L.CheckString(2)
	value := L.CheckAny(3)

	updateResource(r, index, value)

	return 0
}

type ResourceAction func(*Resource) error
