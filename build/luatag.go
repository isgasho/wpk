package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/schwarzlichtbezirk/wpk"
	"github.com/yuin/gopher-lua"
)

const TagMT = "tag"

type LuaTag struct {
	wpk.Tag
}

func RegTag(ls *lua.LState) {
	var mt = ls.NewTypeMetatable(TagMT)
	ls.SetGlobal(TagMT, mt)
	// static attributes
	ls.SetField(mt, "newhex", ls.NewFunction(NewTagHex))
	ls.SetField(mt, "newbase64", ls.NewFunction(NewTagBase64))
	ls.SetField(mt, "newstring", ls.NewFunction(NewTagString))
	ls.SetField(mt, "newbool", ls.NewFunction(NewTagBool))
	ls.SetField(mt, "newuint16", ls.NewFunction(NewTagUint16))
	ls.SetField(mt, "newuint32", ls.NewFunction(NewTagUint32))
	ls.SetField(mt, "newuint64", ls.NewFunction(NewTagUint64))
	ls.SetField(mt, "newnumber", ls.NewFunction(NewTagNumber))
	// methods
	ls.SetField(mt, "__index", ls.NewFunction(getter_tag))
	ls.SetField(mt, "__newindex", ls.NewFunction(setter_tag))
	ls.SetField(mt, "__tostring", ls.NewFunction(string_tag))
	ls.SetField(mt, "__len", ls.NewFunction(len_tag))
	for name, f := range methods_tag {
		ls.SetField(mt, name, ls.NewFunction(f))
	}
	for i, p := range properties_tag {
		ls.SetField(mt, p.name, lua.LNumber(i))
	}
}

func PushTag(ls *lua.LState, v *LuaTag) {
	var ud = ls.NewUserData()
	ud.Value = v
	ls.SetMetatable(ud, ls.GetTypeMetatable(TagMT))
	ls.Push(ud)
}

// Construct LuaTag by given hexadecimal data representation.
func NewTagHex(ls *lua.LState) int {
	var val = ls.CheckString(1)
	var ds, _ = hex.DecodeString(val)
	PushTag(ls, &LuaTag{ds})
	return 1
}

// Construct LuaTag by given base-64 data representation.
func NewTagBase64(ls *lua.LState) int {
	var val = ls.CheckString(1)
	var ds, _ = base64.StdEncoding.DecodeString(val)
	PushTag(ls, &LuaTag{ds})
	return 1
}

// Construct LuaTag by given string.
func NewTagString(ls *lua.LState) int {
	var val = ls.CheckString(1)
	PushTag(ls, &LuaTag{wpk.TagString(val)})
	return 1
}

// Construct LuaTag by given boolean value.
func NewTagBool(ls *lua.LState) int {
	var val = ls.CheckBool(1)
	PushTag(ls, &LuaTag{wpk.TagBool(val)})
	return 1
}

// Construct LuaTag by given uint16 value.
func NewTagUint16(ls *lua.LState) int {
	var val = uint16(ls.CheckInt(1))
	PushTag(ls, &LuaTag{wpk.TagUint16(val)})
	return 1
}

// Construct LuaTag by given uint32 value.
func NewTagUint32(ls *lua.LState) int {
	var val = uint32(ls.CheckInt(1))
	PushTag(ls, &LuaTag{wpk.TagUint32(val)})
	return 1
}

// Construct LuaTag by given uint64 value.
func NewTagUint64(ls *lua.LState) int {
	var val = uint64(ls.CheckInt(1))
	PushTag(ls, &LuaTag{wpk.TagUint64(val)})
	return 1
}

// Construct LuaTag by given number value.
func NewTagNumber(ls *lua.LState) int {
	var val = float64(ls.CheckNumber(1))
	PushTag(ls, &LuaTag{wpk.TagNumber(val)})
	return 1
}

// Checks whether the lua argument with given number is
// a *LUserData with *LuaTag and returns this *LuaTag.
func CheckTag(ls *lua.LState, arg int) *LuaTag {
	if v, ok := ls.CheckUserData(arg).Value.(*LuaTag); ok {
		return v
	}
	ls.ArgError(arg, TagMT+" object required")
	return nil
}

func getter_tag(ls *lua.LState) int {
	var mt = ls.GetMetatable(ls.Get(1))
	var val = ls.GetField(mt, ls.CheckString(2))
	switch val := val.(type) {
	case *lua.LFunction:
		ls.Push(val)
		return 1
	case lua.LNumber:
		var l = &properties_tag[int(val)]
		if l.getter == nil {
			ls.RaiseError("no getter \"%s\" of class \"%s\" defined", l.name, TagMT)
			return 0
		}
		ls.Remove(2) // remove getter name
		return l.getter(ls)
	default:
		ls.Push(lua.LNil)
		return 1
	}
}

func setter_tag(ls *lua.LState) int {
	var mt = ls.GetMetatable(ls.Get(1))
	var val = ls.GetField(mt, ls.CheckString(2))
	switch val := val.(type) {
	case *lua.LFunction:
		ls.Push(val)
		return 1
	case lua.LNumber:
		var l = &properties_tag[int(val)]
		if l.setter == nil {
			ls.RaiseError("no setter \"%s\" of class \"%s\" defined", l.name, TagMT)
			return 0
		}
		ls.Remove(2) // remove setter name
		return l.setter(ls)
	default:
		ls.RaiseError("internal error, wrong pointer type at userdata metatable")
		return 0
	}
}

func string_tag(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	ls.Push(lua.LString(hex.EncodeToString(t.Tag)))
	return 1
}

func len_tag(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	ls.Push(lua.LNumber(len(t.Tag)))
	return 1
}

var properties_tag = []struct {
	name   string
	getter lua.LGFunction // getters always must return 1 value
	setter lua.LGFunction // setters always must return no values
}{
	{"hex", gethex, sethex},
	{"base64", getbase64, setbase64},
	{"string", getstring, setstring},
	{"bool", getbool, setbool},
	{"uint16", getuint16, setuint16},
	{"uint32", getuint32, setuint32},
	{"uint64", getuint64, setuint64},
	{"number", getnumber, setnumber},
}

var methods_tag = map[string]lua.LGFunction{
	// no methods
}

func gethex(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	ls.Push(lua.LString(hex.EncodeToString(t.Tag)))
	return 1
}

func sethex(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = ls.CheckString(2)
	var ds, _ = hex.DecodeString(val)
	t.Tag = ds
	return 0
}

func getbase64(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	ls.Push(lua.LString(base64.StdEncoding.EncodeToString(t.Tag)))
	return 1
}

func setbase64(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = ls.CheckString(2)
	var ds, _ = base64.StdEncoding.DecodeString(val)
	t.Tag = ds
	return 0
}

func getstring(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	if val, ok := t.String(); ok {
		ls.Push(lua.LString(val))
		return 1
	}
	return 0
}

func setstring(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = ls.CheckString(2)
	t.Tag = wpk.TagString(val)
	return 0
}

func getbool(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	if val, ok := t.Bool(); ok {
		ls.Push(lua.LBool(val))
		return 1
	}
	return 0
}

func setbool(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = ls.CheckBool(2)
	t.Tag = wpk.TagBool(val)
	return 0
}

func getuint16(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	if val, ok := t.Uint16(); ok {
		ls.Push(lua.LNumber(val))
		return 1
	}
	return 0
}

func setuint16(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = uint16(ls.CheckInt(2))
	t.Tag = wpk.TagUint16(val)
	return 0
}

func getuint32(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	if val, ok := t.Uint32(); ok {
		ls.Push(lua.LNumber(val))
		return 1
	}
	return 0
}

func setuint32(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = uint32(ls.CheckInt(2))
	t.Tag = wpk.TagUint32(val)
	return 0
}

func getuint64(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	if val, ok := t.Uint64(); ok {
		ls.Push(lua.LNumber(val))
		return 1
	}
	return 0
}

func setuint64(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = uint64(ls.CheckInt(2))
	t.Tag = wpk.TagUint64(val)
	return 0
}

func getnumber(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	if val, ok := t.Number(); ok {
		ls.Push(lua.LNumber(val))
		return 1
	}
	return 0
}

func setnumber(ls *lua.LState) int {
	var t = CheckTag(ls, 1)
	var val = float64(ls.CheckNumber(2))
	t.Tag = wpk.TagNumber(val)
	return 0
}

// Helps convert Lua-table string keys to associated uint16 ID values.
var NameAid = map[string]uint16{
	"fid":     wpk.AID_FID,
	"name":    wpk.AID_name,
	"created": wpk.AID_created,
	"crt":     wpk.AID_created,

	"crc32":     wpk.AID_CRC32C,
	"crc32ieee": wpk.AID_CRC32IEEE,
	"crc32c":    wpk.AID_CRC32C,
	"crc32k":    wpk.AID_CRC32K,
	"crc64":     wpk.AID_CRC64ISO,
	"crc64iso":  wpk.AID_CRC64ISO,

	"md5":    wpk.AID_MD5,
	"sha1":   wpk.AID_SHA1,
	"sha224": wpk.AID_SHA224,
	"sha256": wpk.AID_SHA256,
	"sha384": wpk.AID_SHA384,
	"sha512": wpk.AID_SHA512,

	"mime":     wpk.AID_mime,
	"keywords": wpk.AID_keywords,
	"category": wpk.AID_category,
	"version":  wpk.AID_version,
	"author":   wpk.AID_author,
	"comment":  wpk.AID_comment,
}

type ErrKeyUndef struct {
	TagKey string
}

func (e *ErrKeyUndef) Error() string {
	return fmt.Sprintf("tag key '%s' is undefined", e.TagKey)
}

var (
	ErrBadTagKey = errors.New("tag key type is not number or string")
	ErrBadTagVal = errors.New("tag value type is not string or boolean or 'tag' userdata")
)

// Convert LValue to uint16 tag identifier. Numbers converts explicitly,
// strings converts to uint16 values wich they presents.
// Error returns on any other case.
func ValueToAid(k lua.LValue) (aid uint16, err error) {
	if n, ok := k.(lua.LNumber); ok {
		aid = uint16(n)
	} else if name, ok := k.(lua.LString); ok {
		if n, ok := NameAid[string(name)]; ok {
			aid = n
		} else {
			err = &ErrKeyUndef{string(name)}
			return
		}
	} else {
		err = ErrBadTagKey
		return
	}
	return
}

// Convert Lua-table to Tagset. Lua-table keys can be number identifiers
// or string names associated ID values. Lua-table values can be strings,
// boolean or "tag" userdata values. Numbers can not be passed to table
// to prevent ambiguous type representation.
func TableToTagset(lt *lua.LTable) (ts wpk.Tagset, err error) {
	ts = wpk.Tagset{}
	lt.ForEach(func(k lua.LValue, v lua.LValue) {
		var (
			aid uint16
			tag wpk.Tag
		)

		if aid, err = ValueToAid(k); err != nil {
			return
		}

		if val, ok := v.(lua.LString); ok {
			tag = wpk.TagString(string(val))
		} else if val, ok := v.(lua.LBool); ok {
			tag = wpk.TagBool(bool(val))
		} else if ud, ok := v.(*lua.LUserData); ok {
			if val, ok := ud.Value.(*LuaTag); ok {
				tag = val.Tag
			} else {
				err = ErrBadTagVal
				return
			}
		} else {
			err = ErrBadTagVal
			return
		}

		ts[aid] = tag
	})
	return
}

// The End.