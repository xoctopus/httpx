package scanner

import (
	"encoding"
	"reflect"
)

var (
	tTextUnmarshaler = reflect.TypeFor[encoding.TextUnmarshaler]()
	tBytes           = reflect.TypeFor[[]byte]()
)

func Implements(typ, tar reflect.Type) bool {
	return typ.Implements(tar) || reflect.PointerTo(typ).Implements(tar)
}

func CanUnmarshalByString(typ reflect.Type) bool {
	if Implements(typ, tTextUnmarshaler) || typ == tBytes {
		return true
	}
	switch typ.Kind() {
	case reflect.String:
		return true
	case reflect.Pointer:
		return CanUnmarshalByString(typ.Elem())
	default:
		return false
	}
}
