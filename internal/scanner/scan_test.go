package scanner_test

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/xoctopus/x/slicex"
	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/scanner"
)

func Test_X(t *testing.T) {
	d, _ := json.Marshal([]byte("1111"))
	t.Log(string(d))

	b := []byte{}
	json.Unmarshal(d, &b)
	t.Log(string(b))

	type X struct {
		V1 int `json:"v1,string,format:bin"`
		V2 int `json:"v2,string,format:oct"`
		V3 int `json:"v3,string,format:hex"`
	}
	var v = X{3, 9, 17}
	_ = v
	// json.Marshal(v.V1)  -> expect "0b11"
	// json.Marshal(v.V2)  -> expect "0o11"
	// json.Marshal(v.V3)  -> expect "0x11"
}

func TestScan(t *testing.T) {
	t.Run("Scan", func(t *testing.T) {
		ExpectPanic[error](t, func() {
			scanner.Structs.Scan(reflect.TypeFor[*Testdata]())
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
		lookup, _ := s.Lookup("path", "orgID")
		Expect(t, lookup.FieldName, Equal("OrgID"))
	})

	t.Run("Field.GetOrNewAt", func(t *testing.T) {
		s, _ := scanner.Structs.Scan(reflect.TypeFor[Testdata]())

		q0, _ := s.Lookup("query", "q0")
		direct, _ := s.Lookup("query", "direct")

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
}
