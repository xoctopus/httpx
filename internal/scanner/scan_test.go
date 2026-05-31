package scanner_test

import (
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/xoctopus/x/slicex"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/jsonv2/json"
	"github.com/xoctopus/httpx/internal/scanner"
)

// Previewing for json/v2 format feature
func Test_X(t *testing.T) {
	d, _ := json.Marshal([]byte("1111"))
	Expect(t, d, Equal([]byte(`"MTExMQ=="`)))

	b := []byte{}
	json.Unmarshal(d, &b)
	Expect(t, b, Equal([]byte(`1111`)))

	// json/v2 format features
	// will support more format options for variant types and user-defined format options
	type X struct {
		Int1 int `json:"int1,string,format:bin"` // "0b11" string option effect json.StringifyNumbers option
		Int2 int `json:"int2,string,format:oct"` // "0o11"
		Int3 int `json:"int3,string,format:hex"` // "0x11"

		Time1 time.Time `json:"time1,format:rfc3339"`
		Time2 time.Time `json:"time2,format:dateonly"`
		Time3 time.Time `json:"time3,format:timeonly"`
		// nolint:govet force ignore govet linting
		// Time4 time.Time `json:"time4,format:'2006-01-02 15:04:05'"`

		Duration1 time.Duration `json:",format:sec"`
		Duration2 time.Duration `json:",format:nano"`
		Duration3 time.Duration `json:",format:iso8601"` // PT1H30M 1h30m

		Bytes1 []byte `json:"bytes1,format:base64"`    // default same as v1
		Bytes2 []byte `json:"bytes2,format:base64url"` // for url
		Bytes3 []byte `json:"bytes3,format:hex"`       // abcdef123567890
		Bytes4 []byte `json:"bytes4,format:array"`     // [1,2,3,128]

		ArrayAsNull   []int          `json:"arrayAsNull,format:emitnull"`    // null
		ArrayAsEmpty  []int          `json:"arrayAsEmpty,format:emitempty"`  // []
		ObjectAsNull  map[int]string `json:"objectAsNull,format:emitenull"`  // null
		ObjectAsEmpty map[int]string `json:"objectAsEmpty,format:emitempty"` // {}
	}
	var v = X{}
	_ = v
}

func TestScan(t *testing.T) {
	t.Run("Scan", func(t *testing.T) {
		ExpectPanic[error](t, func() {
			scanner.Structs.Scan(reflect.TypeFor[*int]())
		}, ErrorContains("invalid input type"))

		s, err := scanner.Structs.Scan(reflect.TypeFor[Testdata]())
		Expect(t, err, Succeed())
		Expect(t, s, NotBeNil[*scanner.Struct]())

		fieldsInCookie := slicex.M(
			slices.Collect(s.FieldsInCookie()),
			func(f *scanner.Field) string { return f.Name },
		)
		Expect(t, fieldsInCookie, EquivalentSlice([]string{"token", "userdata"}))

		fieldsInHeader := slicex.M(
			slices.Collect(s.FieldsInHeader()),
			func(f *scanner.Field) string { return f.Name },
		)
		Expect(t, fieldsInHeader, EquivalentSlice([]string{"k1", "k2"}))

		fieldsInPath := slicex.M(
			slices.Collect(s.FieldsInPath()),
			func(f *scanner.Field) string { return f.Name },
		)
		Expect(t, fieldsInPath, EquivalentSlice([]string{"orgID", "userID"}))

		fieldsInQuery := slicex.M(
			slices.Collect(s.FieldsInQuery()),
			func(f *scanner.Field) string { return f.Name },
		)
		Expect(t, fieldsInQuery, EquivalentSlice([]string{
			"q0", "q1", "direct",
			"r1.lt", "r1.lte", "r1.gt", "r1.gte",
			"r2.lt", "r2.lte", "r2.gt", "r2.gte",
		}))

		Expect(t, s.Len(), Equal(18))

		flattened := slicex.M(
			slices.Collect(s.Range),
			func(f *scanner.Field) string { return f.Name },
		)
		Expect(t, flattened, EquivalentSlice([]string{
			"token", "k1", "orgID", "q0", "userdata", "k2", "userID", "q1", "body", "direct",
			"r1.lt", "r1.lte", "r1.gt", "r1.gte",
			"r2.lt", "r2.lte", "r2.gt", "r2.gte",
		}))

		inlined, _ := s.Inlined()
		Expect(t, inlined.FieldName, Equal("Inlined"))
		inPathOrgID, _ := s.LookupIn("path", "orgID")
		Expect(t, inPathOrgID.FieldName, Equal("OrgID"))

		inCookieToken, _ := s.Lookup("token")
		Expect(t, inCookieToken.FieldName, Equal("Token"))
	})

	t.Run("Field.GetOrNewAt", func(t *testing.T) {
		s, _ := scanner.Structs.Scan(reflect.TypeFor[Testdata]())

		q0, _ := s.LookupIn("query", "q0")
		direct, _ := s.LookupIn("query", "direct")

		rv := reflect.ValueOf(Testdata{InQuery: &InQuery{Q0: 100}, Direct: 101})
		v1 := q0.GetOrNewAt(rv)
		v2 := direct.GetOrNewAt(rv)
		Expect(t, v1.Interface(), Equal[any](100))
		Expect(t, v2.Interface(), Equal[any](101))

		rv = reflect.New(reflect.TypeFor[Testdata]()).Elem()
		v1 = q0.GetOrNewAt(rv)
		v2 = direct.GetOrNewAt(rv)
		Expect(t, v1.Interface(), Equal[any](0))
		Expect(t, v2.Interface(), Equal[any](0))
	})

	t.Run("Field.PatchOptions", func(t *testing.T) {
		s, _ := scanner.Structs.Scan(reflect.TypeFor[Testdata]())
		f, _ := s.Lookup("k1")

		// f has string option
		next := f.PatchOptions(nil)
		has, _ := json.GetOption(next, json.StringifyNumbers)
		Expect(t, has, BeTrue())
	})
}
