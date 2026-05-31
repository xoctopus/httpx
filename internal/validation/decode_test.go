package validation_test

import (
	"bytes"
	"errors"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/xoctopus/x/misc/must"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/jsonv2/jsontext"
	"github.com/xoctopus/httpx/internal/scanner"
	"github.com/xoctopus/httpx/internal/validation"
	"github.com/xoctopus/httpx/internal/validation/rule"
	vax "github.com/xoctopus/httpx/internal/validation/va"
	_ "github.com/xoctopus/httpx/pkg/validation/regex"
	"github.com/xoctopus/httpx/pkg/validation/validators"
	_ "github.com/xoctopus/httpx/pkg/validation/validators"
)

type EthAddress string

func (EthAddress) ValidationTag() string {
	return "@test-eth-address"
}

type InvalidValidationTag string

func (InvalidValidationTag) ValidationTag() string {
	return "invalid-tag"
}

type UnregisteredTag string

func (UnregisteredTag) ValidationTag() string {
	return "@unregistered-tag"
}

type JsonUnmarshalerFrom struct{ V any }

func (x *JsonUnmarshalerFrom) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	v, _ := dec.ReadValue()
	_ = string(v)

	x.V = EthAddressValue
	return nil
}

func (x JsonUnmarshalerFrom) MarshalJSON() ([]byte, error) {
	return EthAddressValueRaw, nil
}

type JsonUnmarshaler struct{ V any }

func (x *JsonUnmarshaler) UnmarshalJSON([]byte) error {
	x.V = float32(3.3)
	return nil
}

func (x JsonUnmarshaler) MarshalJSON() ([]byte, error) {
	return []byte("3.3"), nil
}

type TextUnmarshaler struct{ V any }

func (x *TextUnmarshaler) UnmarshalText([]byte) error {
	x.V = int64(5)
	return nil
}

func (x TextUnmarshaler) MarshalText() ([]byte, error) {
	return []byte("5"), nil
}

func init() {
	validation.Register(validators.NewRegexpProvider(
		`^0x[0-9a-fA-F]{40}$`,
		"test-eth-address",
		"标准eth地址(不区分大小写)",
	))
	validation.Register(validators.NewUserDefinedProvider(
		"test-must-failed-rule",
		func(_ string) error { return MustFailedRuleErr },
	))
	validation.Register(validators.NewUserDefinedProvider(
		"test-must-succeed-rule",
		func(_ string) error { return nil },
	))
}

var (
	EthAddressValue    = EthAddress("0x71C7656EC7ab88b098defB751B7401B5f6d1476B")
	EthAddressValueRaw = []byte(strconv.Quote(string(EthAddressValue)))

	MockValidationErr = errors.New("mock-validation-error")
	MustFailedRuleErr = errors.New("test-must-failed-rule-error")
)

func NewMockValidator(needPass bool) validation.Validator {
	return &MockValidator{pass: needPass}
}

type MockValidator struct {
	pass bool
}

func (va MockValidator) String() string {
	return "test-mock"
}

func (va MockValidator) Validate(_ []byte) error {
	if va.pass {
		return nil
	}
	return MockValidationErr
}

type Embedded struct {
	Int                 int                 `json:"int,omitzero"`
	Uint                uint                `json:"uint,omitzero"`
	Float               float32             `json:"float,omitzero"`
	Bool                bool                `json:"bool,omitzero"`
	Map                 map[int]string      `json:"map,omitzero"`
	Bytes               []byte              `json:"bytes,omitzero"`
	Slice               []float64           `json:"array,omitzero"`
	Any                 any                 `json:"any,omitzero"`
	Pointer             *string             `json:"pointer,omitzero"`
	JsonUnmarshalerFrom JsonUnmarshalerFrom `json:"jsonUnmarshalerFrom,omitzero"`
	JsonUnmarshaler     JsonUnmarshaler     `json:"jsonUnmarshaler,omitzero"`
	TextUnmarshaler     TextUnmarshaler     `json:"textUnmarshaler,omitzero"`
}

type Testdata struct {
	Wrap       int            `json:"Wrap"`
	EthAddress EthAddress     `json:"CustomTagValidator"`
	Pointer    *int           `json:"Pointer"`
	Primitive  float32        `json:"Primitive,string"`
	Map        map[string]int `json:"Map" validate:"@map<@string,@int>[2]"`
	Any        any            `json:"Any"`
	Array      [3]string      `json:"Array"`
	Slice      []any          `json:"Slice"`
	Struct     *Embedded      `json:"Struct"`
}

func TestUnmarshalDecode(t *testing.T) {
	v := &Testdata{}
	rv := reflect.ValueOf(v)
	re := rv.Elem()

	s, _ := scanner.Structs.Scan(rv.Type())
	Expect(t, s, NotBeNil[*scanner.Struct]())

	fieldForTestCase := func(t *testing.T) *scanner.Field {
		parts := strings.Split(t.Name(), "/")
		f, ok := s.Lookup(parts[len(parts)-1])
		must.BeTrueF(ok, "missing testing case field: %s", t.Name())
		return f
	}

	t.Run("Wrap", func(t *testing.T) {
		f := fieldForTestCase(t)
		w := scanner.WrapField(rv.Elem().Field(0).Addr(), f)
		d := jsontext.NewDecoder(bytes.NewBufferString("100"))

		Expect(t, validation.UnmarshalDecode(d, w), Succeed())
		Expect(t, v.Wrap, Equal(100))
	})

	t.Run("CustomTagValidator", func(t *testing.T) {
		f := fieldForTestCase(t)
		w := scanner.WrapField(rv.Elem().Field(1).Addr(), f)

		t.Run("FromField", func(t *testing.T) {
			d1 := jsontext.NewDecoder(bytes.NewReader(EthAddressValueRaw))
			Expect(t, validation.UnmarshalDecode(d1, w), Succeed())
			Expect(t, v.EthAddress, Equal(EthAddressValue))

			d2 := jsontext.NewDecoder(bytes.NewBufferString(`"0xInvalid"`))
			err := validation.UnmarshalDecode(d2, w)
			Expect(t, err, IsCodeError(validation.ERROR__NOT_MATCH_REGEXP))
		})
		t.Run("FromValue", func(t *testing.T) {
			v1 := new(EthAddress)
			d1 := jsontext.NewDecoder(bytes.NewBuffer(EthAddressValueRaw))
			Expect(t, validation.UnmarshalDecode(d1, v1), Succeed())
			Expect(t, *v1, Equal(EthAddressValue))

			v2 := new(EthAddress)
			d2 := jsontext.NewDecoder(bytes.NewBuffer(EthAddressValueRaw))
			Expect(t, validation.UnmarshalDecode(d2, reflect.ValueOf(v2)), Succeed())
			Expect(t, *v2, Equal(EthAddressValue))
		})
		t.Run("FailedCompiling", func(t *testing.T) {
			Expect(
				t,
				validation.UnmarshalDecode(
					jsontext.NewDecoder(bytes.NewReader([]byte("any"))),
					new(InvalidValidationTag),
				),
				IsCodeError(rule.ERROR__INVALID_RULE_LEADER),
			)
		})
		t.Run("UnregisteredValidationRule", func(t *testing.T) {
			Expect(
				t,
				validation.UnmarshalDecode(
					jsontext.NewDecoder(bytes.NewReader([]byte("any"))),
					new(UnregisteredTag),
				),
				IsCodeError(validation.ERROR__UNREGISTERED_RULE),
			)
		})
	})

	t.Run("Pointer", func(t *testing.T) {
		f := fieldForTestCase(t)
		t.Run("Null", func(t *testing.T) {
			d := jsontext.NewDecoder(bytes.NewBufferString("null"))
			u := validation.Pointer(f.GetOrNewAt(re), nil)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, v.Pointer, BeNil[*int]())
			t.Run("FailedReadingValue", func(t *testing.T) {
				d = jsontext.NewDecoder(bytes.NewBufferString("nul"))
				u = validation.Pointer(f.GetOrNewAt(re), nil)
				Expect(t, u.UnmarshalDecode(d), Failed())
			})
			t.Run("FailedValidating", func(t *testing.T) {
				d = jsontext.NewDecoder(bytes.NewBufferString("null"))
				u = validation.Pointer(f.GetOrNewAt(re), NewMockValidator(false))
				Expect(t, u.UnmarshalDecode(d), IsError(MockValidationErr))
			})
		})
		t.Run("NonNull", func(t *testing.T) {
			d := jsontext.NewDecoder(bytes.NewBufferString("100"))
			u := validation.Pointer(f.GetOrNewAt(re), nil)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, *v.Pointer, Equal(100))
			t.Run("FailedDecoding", func(t *testing.T) {
				d = jsontext.NewDecoder(bytes.NewBufferString("100"))
				u = validation.Pointer(f.GetOrNewAt(re), NewMockValidator(false))
				Expect(t, u.UnmarshalDecode(d), IsError(MockValidationErr))
			})

		})
		t.Run("Raw", func(t *testing.T) {
			x := new(int)
			d := jsontext.NewDecoder(bytes.NewBufferString("100"))
			u := validation.Pointer(reflect.ValueOf(x), nil)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, *v.Pointer, Equal(100))
		})
	})

	t.Run("Primitive", func(t *testing.T) {
		f := fieldForTestCase(t)

		d := jsontext.NewDecoder(bytes.NewReader([]byte(`"100.01"`)))
		u := validation.Basic(f.GetOrNewAt(re), nil)
		Expect(t, u.UnmarshalDecode(d), Succeed())
		Expect(t, v.Primitive, Equal[float32](100.01))

		t.Run("FailedToUnmarshal", func(t *testing.T) {
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`"abc"`)))
			u = validation.Basic(f.GetOrNewAt(re), nil)
			Expect(t, u.UnmarshalDecode(d), Failed())
		})
		t.Run("FirstElementInArray", func(t *testing.T) {
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`["100.02","200"]`)))
			u = validation.Basic(f.GetOrNewAt(re), nil)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, v.Primitive, Equal[float32](100.02))
			t.Run("MustNonemptyArray", func(t *testing.T) {
				d = jsontext.NewDecoder(bytes.NewReader([]byte(`[]`)))
				u = validation.Basic(f.GetOrNewAt(re), nil)
				Expect(t, u.UnmarshalDecode(d), Failed())
			})
		})
	})

	t.Run("Map", func(t *testing.T) {
		f := fieldForTestCase(t)
		va, _ := validation.NewFromStructField(f)

		d := jsontext.NewDecoder(bytes.NewBufferString(`{"a":1,"b":2}`))
		u := validation.Map(f.GetOrNewAt(re), va)
		Expect(t, u.UnmarshalDecode(d), Succeed())
		Expect(t, v.Map, Equal(map[string]int{"a": 1, "b": 2}))

		t.Run("Null", func(t *testing.T) {
			d = jsontext.NewDecoder(bytes.NewReader([]byte("null")))
			u = validation.Map(f.GetOrNewAt(re), NewMockValidator(true))
			Expect(t, u.UnmarshalDecode(d), Succeed())
			t.Run("FailedReadingValue", func(t *testing.T) {
				d = jsontext.NewDecoder(bytes.NewReader([]byte("null")))
				u = validation.Map(f.GetOrNewAt(re), NewMockValidator(false))
				Expect(t, u.UnmarshalDecode(d), IsError(MockValidationErr))
			})
		})
		t.Run("Object", func(t *testing.T) {
			v.Map = nil
			t.Run("InvalidKeyRule", func(t *testing.T) {
				va := &validators.Map{}
				va.SetKeyRule(rule.MustCompile("@invalid"))
				d = jsontext.NewDecoder(bytes.NewReader([]byte("{}")))
				u = validation.Map(f.GetOrNewAt(re), va)
				Expect(t, u.UnmarshalDecode(d), Failed())
			})
			t.Run("InvalidElemRule", func(t *testing.T) {
				va := &validators.Map{}
				va.SetElemRule(rule.MustCompile("@invalid"))
				d = jsontext.NewDecoder(bytes.NewReader([]byte("{}")))
				u = validation.Map(f.GetOrNewAt(re), va)
				Expect(t, u.UnmarshalDecode(d), Failed())
			})
			t.Run("DuplicateKey", func(t *testing.T) {
				va := &validators.Map{}
				d = jsontext.NewDecoder(bytes.NewReader([]byte(`{"a":1,"a":2}`)))
				u = validation.Map(f.GetOrNewAt(re), va)
				err, ok := errors.AsType[*jsontext.SyntacticError](u.UnmarshalDecode(d))
				Expect(t, ok, BeTrue())
				Expect(t, err.Err, IsError(jsontext.ErrDuplicateName))
			})
			t.Run("FailedToDecodeKey", func(t *testing.T) {
				va := &validators.Map{}
				va.SetKeyRule(rule.MustCompile("@test-must-failed-rule"))
				va.SetElemRule(rule.MustCompile("@test-must-succeed-rule"))

				d = jsontext.NewDecoder(bytes.NewReader([]byte(`{"x":1}`)))
				u = validation.Map(f.GetOrNewAt(re), va)
				Expect(t, u.UnmarshalDecode(d), IsError(MustFailedRuleErr))
			})
			t.Run("FailedToDecodeElem", func(t *testing.T) {
				va := &validators.Map{}
				va.SetKeyRule(rule.MustCompile("@test-must-succeed-rule"))

				d = jsontext.NewDecoder(bytes.NewReader([]byte(`{"x":"abc"}`)))
				u = validation.Map(f.GetOrNewAt(re), va)
				Expect(t, u.UnmarshalDecode(d), Failed())
			})
			t.Run("PostValidating", func(t *testing.T) {
				v.Map = nil
				d = jsontext.NewDecoder(bytes.NewReader([]byte(`{"x":100,"y":101,"z":103}`)))
				u = validation.Map(f.GetOrNewAt(re), va)
				Expect(t, u.UnmarshalDecode(d), IsCodeError(vax.ERROR__OUT_OF_LENGTH))

				v.Map = nil
				d = jsontext.NewDecoder(bytes.NewReader([]byte(`{"x":100,"y":101}`)))
				u = validation.Map(f.GetOrNewAt(re), va)
				Expect(t, u.UnmarshalDecode(d), Succeed())
			})
			t.Run("FailedPostValidating", func(t *testing.T) {
			})
		})
		t.Run("SemanticError", func(n *testing.T) {
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`"x"`)))
			u = validation.Map(f.GetOrNewAt(re), va)
			err, ok := errors.AsType[*json.SemanticError](u.UnmarshalDecode(d))
			Expect(t, ok, BeTrue())
			Expect(t, err.JSONKind, Equal(jsontext.STRING))
		})
	})

	t.Run("Any", func(t *testing.T) {
		f := fieldForTestCase(t)
		va, _ := validation.NewFromStructField(f)

		d := jsontext.NewDecoder(bytes.NewBufferString(`100.001`))
		u := validation.Any(f.GetOrNewAt(re), va)
		Expect(t, u.UnmarshalDecode(d), Succeed())
		Expect(t, v.Any, Equal[any](100.001))

		t.Run("Null", func(t *testing.T) {
			x := new(any)
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`null`)))
			u = validation.Any(reflect.ValueOf(x).Elem(), nil)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, *x, BeNil[any]())
		})
	})

	t.Run("Array", func(t *testing.T) {
		f := fieldForTestCase(t)
		va, _ := validation.NewFromStructField(f)

		d := jsontext.NewDecoder(bytes.NewBufferString(`["1","2","3"]`))
		u := validation.Array(f.GetOrNewAt(re), va)
		Expect(t, u.UnmarshalDecode(d), Succeed())
		Expect(t, v.Array, Equal([3]string{"1", "2", "3"}))

		t.Run("Null", func(t *testing.T) {
			v.Array = [3]string{}
			d = jsontext.NewDecoder(bytes.NewBufferString(`null`))
			u = validation.Array(f.GetOrNewAt(re), nil)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, v.Array, Equal([3]string{}))

			t.Run("Failed", func(t *testing.T) {
				v.Array = [3]string{}
				d = jsontext.NewDecoder(bytes.NewBufferString(`null`))
				u = validation.Array(f.GetOrNewAt(re), NewMockValidator(false))
				Expect(t, u.UnmarshalDecode(d), IsError(MockValidationErr))
			})
		})

		t.Run("EmptyArray", func(t *testing.T) {
			x := new([0]int)
			u = validation.Pointer(reflect.ValueOf(x), nil)

			d1 := jsontext.NewDecoder(bytes.NewReader([]byte(`[]`)))
			Expect(t, u.UnmarshalDecode(d1), Succeed())
			d2 := jsontext.NewDecoder(bytes.NewReader([]byte(`null`)))
			Expect(t, u.UnmarshalDecode(d2), Succeed())
		})
		t.Run("EmptySlice", func(t *testing.T) {
			x := new([]int)
			u = validation.Pointer(reflect.ValueOf(x), nil)

			d1 := jsontext.NewDecoder(bytes.NewReader([]byte(`[]`)))
			Expect(t, u.UnmarshalDecode(d1), Succeed())
			Expect(t, len(*x), Equal(0))
			Expect(t, cap(*x), Equal(0))

			d2 := jsontext.NewDecoder(bytes.NewReader([]byte(`null`)))
			Expect(t, u.UnmarshalDecode(d2), Succeed())
			Expect(t, len(*x), Equal(0))
			Expect(t, cap(*x), Equal(0))
		})
		t.Run("NonemptySlice", func(t *testing.T) {
			x := new([]int)
			u = validation.Pointer(reflect.ValueOf(x), nil)
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`[1,2,3]`)))
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, len(*x), Equal(3))
		})
		t.Run("FailedToPostValidating", func(t *testing.T) {
			va, err := validation.New(validation.Option{
				Type: reflect.TypeFor[[]int](),
				Rule: rule.MustCompile("@slice<@int>[2]"),
			})
			Expect(t, err, Succeed())
			x := new([]int)
			u = validation.Pointer(reflect.ValueOf(x), va)
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`[1,2,3]`)))
			Expect(t, u.UnmarshalDecode(d), IsCodeError(vax.ERROR__OUT_OF_LENGTH))
		})
		t.Run("AssignableToBytes", func(t *testing.T) {
			x := new(net.IP)
			va, _ := validation.New(validation.Option{
				Type: reflect.TypeFor[net.IP](),
				Rule: rule.MustCompile("@ipv4"),
			})

			d := jsontext.NewDecoder(bytes.NewReader([]byte(`"1.1.1.1"`)))
			u := validation.Value(reflect.ValueOf(x).Elem(), va)
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, x.String(), Equal("1.1.1.1"))
		})
		t.Run("SemanticError", func(t *testing.T) {
			x := new([]int)
			u = validation.Pointer(reflect.ValueOf(x), va)
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`3`)))
			Expect(t, u.UnmarshalDecode(d), Failed())
		})
	})

	t.Run("Struct", func(t *testing.T) {
		f := fieldForTestCase(t)
		va, _ := validation.NewFromStructField(f)

		d := jsontext.NewDecoder(bytes.NewReader([]byte(`{
  "int":1,
  "uint":2,
  "float":3.4,
  "bool":true,
  "map":{"1":"a","2":"b"},
  "bytes":"MTExMQ==",
  "array":[ 1.1, 2.2 ],
  "any":"abc",
  "pointer":"def",
  "jsonUnmarshalerFrom":"0x71C7656EC7ab88b098defB751B7401B5f6d1476B",
  "jsonUnmarshaler":3.3,
  "textUnmarshaler":"5"
}`)))
		u := validation.Value(f.GetOrNewAt(re), va)
		Expect(t, u.UnmarshalDecode(d), Succeed())
		Expect(t, *v.Struct, Equal(Embedded{
			Int:                 1,
			Uint:                2,
			Float:               3.4,
			Bool:                true,
			Map:                 map[int]string{1: "a", 2: "b"},
			Bytes:               []byte("1111"),
			Slice:               []float64{1.1, 2.2},
			Any:                 "abc",
			Pointer:             new("def"),
			JsonUnmarshalerFrom: JsonUnmarshalerFrom{V: EthAddressValue},
			JsonUnmarshaler:     JsonUnmarshaler{V: float32(3.3)},
			TextUnmarshaler:     TextUnmarshaler{V: int64(5)},
		}))

		t.Run("Null", func(t *testing.T) {
			x := new(Embedded)
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`null`)))
			u = validation.Struct(reflect.ValueOf(x).Elem(), NewMockValidator(true))
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, *x, Equal(Embedded{}))
		})
		t.Run("SematicError", func(t *testing.T) {
			x := new(struct{})
			d = jsontext.NewDecoder(bytes.NewReader([]byte(`1`)))
			u = validation.Struct(reflect.ValueOf(x).Elem(), nil)
			Expect(t, u.UnmarshalDecode(d), Failed())
		})
		t.Run("PartialFields", func(t *testing.T) {
			type X struct {
				A int    `json:"a"`
				B string `json:"b" default:"\"abc\""`
			}
			x := new(X)
			d := jsontext.NewDecoder(bytes.NewReader([]byte(`{"a":1}`)))
			u := validation.Struct(reflect.ValueOf(x).Elem(), NewMockValidator(true))
			Expect(t, u.UnmarshalDecode(d), Succeed())
			Expect(t, *x, Equal(X{A: 1, B: "abc"}))
			t.Run("FailedToDecodeDefaults", func(t *testing.T) {
				type X struct {
					A int    `json:"a"`
					B string `json:"b" default:"abc"` // invalid string value
				}
				x := new(X)
				d := jsontext.NewDecoder(bytes.NewReader([]byte(`{"a":1}`)))
				u := validation.Struct(reflect.ValueOf(x).Elem(), NewMockValidator(true))
				Expect(t, validation.IsValidationError(u.UnmarshalDecode(d)), BeTrue())
			})
		})
	})
}
