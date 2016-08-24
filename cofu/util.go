package cofu

import (
	"fmt"
	"github.com/yookoala/realpath"
	"github.com/yuin/gopher-lua"
	"path/filepath"
)

// This code inspired by https://github.com/yuin/gluamapper/blob/master/gluamapper.go
func toGoValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case *lua.LFunction:
		return v
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				keystr := fmt.Sprint(toGoValue(key))
				ret[keystr] = toGoValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, toGoValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v
	}
}

func toString(v lua.LValue) (string, bool) {
	if lv, ok := v.(lua.LString); ok {
		return string(lv), true
	} else {
		return "", false
	}
}

// currentDir returns a directory path that includes lua source file which is executed now.
func currentDir(L *lua.LState) string {
	// same: debug.getinfo(2,'S').source
	var dbg *lua.Debug
	var err error
	var ok bool

	dbg, ok = L.GetStack(1)
	if !ok {
		return ""
	}
	_, err = L.GetInfo("S", dbg, lua.LNil)
	if err != nil {
		panic(err)
	}

	return filepath.Dir(dbg.Source)
}

func basepath(L *lua.LState) string {
	p, err := realpath.Realpath(currentDir(L))
	if err != nil {
		panic(err)
	}

	return p
}
