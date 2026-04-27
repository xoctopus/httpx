package scanner_test

import (
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/scanner"
)

type InBody struct {
	Name    string `json:"name,nocase,strictcase"`
	Data    []byte `json:"data"`
	Ignored any    `name:"-"`
	NoTag   any
	Strings []string
}

type InPath struct {
	OrgID  int `in:"path" name:"orgID"`
	UserID int `in:"path" name:"userID"`
}

type InQuery struct {
	Q0 int    `in:"query" name:"q0,format:bin"`
	Q1 string `in:"query" name:"q1"`

	R1 Ranger[int]     `in:"query" name:"r1,embedded"`
	R2 Ranger[float64] `in:"query" name:"r2,embedded"`
}

type Ranger[T comparable] struct {
	Min  T `in:"query" name:"gt"`
	MinE T `in:"query" name:"gte"`
	Max  T `in:"query" name:"lt"`
	MaxE T `in:"query" name:"lte"`
}

type InHeader struct {
	V1 int    `in:"header" name:"k1"`
	V2 string `in:"header" name:"k2"`
}

type InCookie struct {
	Token string `in:"cookie" name:"token"`
	CookiePayload
}

type CookiePayload struct {
	Userdata string `in:"cookie" name:"userdata"`
}

type Inlined []string

type Unknown struct{}

type Testdata struct {
	// httpx.MethodGet `path:"/org/{orgID}/user/{userID}"`

	Direct int `in:"query" name:"direct"`

	InPath
	*InQuery
	InCookie
	InHeader

	InBody  `in:"body"`
	Inlined // inlined: embedded field and not a struct

	// for coverage
	Unknown `name:",unknown"`
	Ignored any `name:"-"`
}

func TestParseFieldOptions(t *testing.T) {
	tt := reflect.TypeFor[Testdata]()

	sf, _ := tt.FieldByName("NoTag")
	options, ignored, err := scanner.ParseFieldOptions(sf)
	Expect(t, err, Succeed())
	Expect(t, options, Equal(scanner.FieldOptions{Name: "NoTag"}))
	Expect(t, ignored, BeFalse())

	sf, _ = tt.FieldByName("Ignored")
	options, ignored, err = scanner.ParseFieldOptions(sf)
	Expect(t, err, Succeed())
	Expect(t, options, Equal(scanner.FieldOptions{Name: "Ignored"}))
	Expect(t, ignored, BeTrue())

	sf, _ = tt.FieldByName("Name")
	options, ignored, err = scanner.ParseFieldOptions(sf)
	Expect(t, err, ErrorEqual("couldn't have both `nocase` and `strictcase`"))
	Expect(t, ignored, BeFalse())

	sf, _ = tt.FieldByName("Data")
	options, ignored, err = scanner.ParseFieldOptions(sf)
	Expect(t, err, Succeed())
	Expect(t, options, Equal(scanner.FieldOptions{
		Name:    "data",
		HasName: true,
		String:  true,
	}))
	Expect(t, ignored, BeFalse())
	Expect(t, options.StrictCase(), BeFalse())

	sf, _ = tt.FieldByName("InBody")
	options, ignored, err = scanner.ParseFieldOptions(sf)
	Expect(t, err, Succeed())
	Expect(t, options, Equal(scanner.FieldOptions{Name: "body", HasName: true}))
	Expect(t, ignored, BeFalse())

	sf, _ = tt.FieldByName("Q0")
	options, ignored, err = scanner.ParseFieldOptions(sf)
	Expect(t, err, Succeed())
	Expect(t, options, Equal(scanner.FieldOptions{Name: "q0", HasName: true, Format: "bin"}))
	Expect(t, ignored, BeFalse())

	sf, _ = tt.FieldByName("Strings")
	options, ignored, err = scanner.ParseFieldOptions(sf)
	Expect(t, err, Succeed())
	Expect(t, options, Equal(scanner.FieldOptions{Name: "Strings", StringItem: true}))
	Expect(t, ignored, BeFalse())
}
