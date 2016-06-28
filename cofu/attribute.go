package cofu

import (
	"fmt"
	"github.com/yuin/gopher-lua"
	"strconv"
)

var CommonAttributes = []Attribute{
	&StringAttribute{
		Name: "only_if",
	},
	&StringAttribute{
		Name: "not_if",
	},
	&StringAttribute{
		Name: "user",
	},
	&StringAttribute{
		Name: "cwd",
	},
	&NotifiesAttribute{
		Name:    "notifies",
		Default: nil,
	},
	&StringSliceAttribute{
		Name:    "verify",
		Default: nil,
	},
}

type Attribute interface {
	GetName() string
	IsRequired() bool
	HasDefault() bool
	GetDefault() interface{}
	ToGoValue(v lua.LValue) interface{}
}

// StringAttribute
type StringAttribute struct {
	Name     string
	Required bool
	Default  string
	// If DefaultName is true, it uses name as default value.
	DefaultName bool
}

func (attr *StringAttribute) GetName() string {
	return attr.Name
}

func (attr *StringAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *StringAttribute) HasDefault() bool {
	return attr.Default != ""
}

func (attr *StringAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *StringAttribute) IsDefaultName() bool {
	return attr.DefaultName
}

func (attr *StringAttribute) ToGoValue(lv lua.LValue) interface{} {
	v := toGoValue(lv)
	if fv, ok := v.(float64); ok {
		v = fmt.Sprintf("%v", fv)
	}

	return v
}

// StringSliceAttribute
type StringSliceAttribute struct {
	Name     string
	Required bool
	Default  []string
	// If DefaultName is true, it uses name as default value.
	DefaultName bool
}

func (attr *StringSliceAttribute) GetName() string {
	return attr.Name
}

func (attr *StringSliceAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *StringSliceAttribute) HasDefault() bool {
	return attr.Default != nil
}

func (attr *StringSliceAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *StringSliceAttribute) IsDefaultName() bool {
	return attr.DefaultName
}

func (attr *StringSliceAttribute) ToGoValue(v lua.LValue) interface{} {
	gov := toGoValue(v)

	if ss, ok := gov.([]interface{}); ok {
		ret := []string{}
		for _, s := range ss {
			ret = append(ret, s.(string))
		}
		return ret
	} else {
		return gov
	}
}

// MapAttribute
type MapAttribute struct {
	Name     string
	Required bool
	Default  map[string]interface{}
}

func (attr *MapAttribute) GetName() string {
	return attr.Name
}

func (attr *MapAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *MapAttribute) HasDefault() bool {
	return attr.Default != nil
}

func (attr *MapAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *MapAttribute) ToGoValue(v lua.LValue) interface{} {
	return toGoValue(v)
}

// BoolAttribute
type BoolAttribute struct {
	Name     string
	Required bool
	Default  bool
}

func (attr *BoolAttribute) GetName() string {
	return attr.Name
}

func (attr *BoolAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *BoolAttribute) HasDefault() bool {
	return attr.Default
}

func (attr *BoolAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *BoolAttribute) ToGoValue(v lua.LValue) interface{} {
	return toGoValue(v)
}

type ComparableValue interface {
	String() string
	Nil() bool
}

type Integer struct {
	V     int
	IsNil bool
}

func (i *Integer) String() string {
	return fmt.Sprintf("%d", i.V)
}

func (i *Integer) Nil() bool {
	return i.IsNil
}

// IntAttribute
type IntegerAttribute struct {
	Name     string
	Required bool
	Default  *Integer
}

func (attr *IntegerAttribute) GetName() string {
	return attr.Name
}

func (attr *IntegerAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *IntegerAttribute) HasDefault() bool {
	return attr.Default != nil
}

func (attr *IntegerAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *IntegerAttribute) ToGoValue(lv lua.LValue) interface{} {

	if _, ok := lv.(*lua.LNilType); ok {
		// is nil
		return &Integer{
			IsNil: true,
		}
	}

	if n, ok := lv.(lua.LNumber); ok {
		i := &Integer{
			V:     int(float64(n)),
			IsNil: false,
		}

		return i
	}

	if s, ok := lv.(lua.LString); ok {
		n, err := strconv.Atoi(string(s))
		if err != nil {
			panic(err)
		}
		i := &Integer{
			V:     n,
			IsNil: false,
		}

		return i
	}

	panic("'" + attr.Name + "' must be a number")
}

// LFunctionAttribute
type LFunctionAttribute struct {
	Name     string
	Required bool
	Default  *lua.LFunction
}

func (attr *LFunctionAttribute) GetName() string {
	return attr.Name
}

func (attr *LFunctionAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *LFunctionAttribute) HasDefault() bool {
	return attr.Default != nil
}

func (attr *LFunctionAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *LFunctionAttribute) ToGoValue(lv lua.LValue) interface{} {
	v, ok := lv.(*lua.LFunction)
	if !ok {
		panic("'" + attr.Name + "' must be a function")
	}

	return v
}

// NotifiesAttribute
type NotifiesAttribute struct {
	Name     string
	Required bool
	Default  []*Notification
}

func (attr *NotifiesAttribute) GetName() string {
	return attr.Name
}

func (attr *NotifiesAttribute) IsRequired() bool {
	return attr.Required
}

func (attr *NotifiesAttribute) HasDefault() bool {
	return attr.Default != nil
}

func (attr *NotifiesAttribute) GetDefault() interface{} {
	return attr.Default
}

func (attr *NotifiesAttribute) ToGoValue(lv lua.LValue) interface{} {
	// lua code example:
	//   notifies = {"restart", "httpd"}
	//   notifies = {"restart", "service[httpd]", "immediately"}
	//   notifies = {{"restart", "service[httpd]", "immediately"}, {"restart", "service[nginx]"}}

	notifications := []*Notification{}

	v, ok := lv.(*lua.LTable)
	if !ok {
		panic("notifies must be array table")
	}

	maxn := v.MaxN()
	if maxn == 0 { // table
		panic("notifies must be array table")
	} else { // array
		if _, ok := v.RawGetInt(1).(lua.LString); ok {
			// only one notificaton config
			notifications = append(notifications, attr.createNotification(v))
		} else if _, ok := v.RawGetInt(1).(*lua.LTable); ok {
			// multi notification config
			for i := 1; i <= maxn; i++ {
				vt, ok := v.RawGetInt(i).(*lua.LTable)
				if !ok {
					panic("notifies must be array table")
				}
				notifications = append(notifications, attr.createNotification(vt))
			}
		}
	}

	return notifications
}

func (attr *NotifiesAttribute) createNotification(config *lua.LTable) *Notification {
	if lv := config.RawGetInt(1); lv == lua.LNil {
		panic("invalid notification config")
	}
	action := config.RawGetInt(1).String()

	if lv := config.RawGetInt(2); lv == lua.LNil {
		panic("invalid notification config")
	}
	desc := config.RawGetInt(2).String()

	timing := "delayed"
	maxn := config.MaxN()
	if maxn == 3 {
		timing = config.RawGetInt(3).String()
	}

	return &Notification{
		Action:             action,
		TargetResourceDesc: desc,
		Timing:             timing,
	}
}
