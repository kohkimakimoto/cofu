package cofu

import (
	"fmt"
	"github.com/kohkimakimoto/loglv"
	"github.com/yuin/gopher-lua"
	"log"
)

type ResourceType struct {
	Name                     string
	Attributes               []Attribute
	PreAction                ResourceAction
	SetCurrentAttributesFunc ResourceAction
	Actions                  map[string]ResourceAction
	ShowDifferences          ResourceAction
	app                      *App
}

func (resourceType *ResourceType) LGFunction() func(L *lua.LState) int {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		// procedural style
		if L.GetTop() == 2 {
			tb := L.CheckTable(2)
			resourceType.registerResource(L, name, tb)

			return 0
		}

		// DSL style
		L.Push(L.NewFunction(func(L *lua.LState) int {
			tb := L.CheckTable(1)
			resourceType.registerResource(L, name, tb)

			return 0
		}))

		return 1
	}
}

func (resourceType *ResourceType) registerResource(L *lua.LState, name string, attrs *lua.LTable) {
	app := GetApp(L)

	r := NewResource(name, resourceType, app)

	if loglv.IsDebug() {
		log.Printf("    (Debug) compiling resource '%s'", r.Desc())
	}

	// set defaults
	for _, definedAttribute := range resourceType.Attributes {
		if definedAttribute.HasDefault() {
			if loglv.IsDebug() {
				log.Printf("    (Debug) Set default: %s = %s", definedAttribute.GetName(), definedAttribute.GetDefault())
			}

			r.Attributes[definedAttribute.GetName()] = definedAttribute.GetDefault()
		}

		switch a := definedAttribute.(type) {
		case *StringAttribute:
			if a.IsDefaultName() {
				if loglv.IsDebug() {
					log.Printf("    (Debug) Set default: %s = %s", definedAttribute.GetName(), r.Name)
				}
				r.Attributes[definedAttribute.GetName()] = r.Name
			}
		case *StringSliceAttribute:
			if a.IsDefaultName() {
				if loglv.IsDebug() {
					log.Printf("    (Debug) Set default: %s = %s", definedAttribute.GetName(), r.Name)
				}
				r.Attributes[definedAttribute.GetName()] = r.Name
			}
		}
	}

	// set attributes.
	attrs.ForEach(func(key lua.LValue, value lua.LValue) {
		lAttributeName, ok := key.(lua.LString)
		if !ok {
			panic(fmt.Sprintf("'%s' An attribute must be string", r.Desc()))
		}
		attributeName := lAttributeName.String()

		var attribute Attribute
		for _, definedAttribute := range resourceType.Attributes {
			if definedAttribute.GetName() == attributeName {
				attribute = definedAttribute
				break
			}
		}

		if attribute == nil {
			panic(fmt.Sprintf("Invalid attribute name '%s'.", attributeName))
		}

		r.Attributes[attribute.GetName()] = attribute.ToGoValue(value)

		attribute = nil
	})

	// checks required attributes
	for _, definedAttribute := range resourceType.Attributes {
		if definedAttribute.IsRequired() {
			if _, ok := r.Attributes[definedAttribute.GetName()]; !ok {
				panic(fmt.Sprintf("'%s' attribute is required but it is not set.", definedAttribute.GetName()))
			}
		}
	}

	// parse and validate notifies attribute
	if r.GetRawAttribute("notifies") != nil {
		r.Notifications = r.GetRawAttribute("notifies").([]*Notification)
		for _, n := range r.Notifications {
			n.DefinedInResource = r
			if err := n.Validate(); err != nil {
				panic(err)
			}
		}
	}

	// set default diff function it it does not have specific func.
	if resourceType.ShowDifferences == nil {
		resourceType.ShowDifferences = DefaultShowDifferences
	}

	app.RegisterResource(r)
}

type ResourceAction func(*Resource) error
