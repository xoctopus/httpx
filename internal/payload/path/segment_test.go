package path_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/payload/path"
)

func TestSegment(t *testing.T) {
	s := path.NewSegment("")
	Expect(t, s, BeNil[path.Segment]())

	s = path.NewSegment("abc")
	Expect(t, s.ParamString(), Equal("abc"))

	s = path.NewSegment("{userID}")
	Expect(t, s.ParamString(), Equal("{userID}"))
	ns, ok := s.(path.NamedSegment)
	Expect(t, ok, BeTrue())
	Expect(t, ns.ParamName(), Equal("userID"))
	Expect(t, ns.Multiple(), BeFalse())

	s = path.NewSegment("{users...}")
	Expect(t, s.ParamString(), Equal("{users...}"))
	ns, ok = s.(path.NamedSegment)
	Expect(t, ok, BeTrue())
	Expect(t, ns.ParamName(), Equal("users"))
	Expect(t, ns.Multiple(), BeTrue())

	ss := path.ParseSegments("/v1/user/{userID}")
	Expect(t, len(ss), Equal(3))
	Expect(t, ss.ParamString(), Equal("/v1/user/{userID}"))

	values, err := ss.PathValues("/v1/user/100")
	Expect(t, err, Succeed())
	Expect(t, values, Equal(path.Values{"userID": "100"}))

	_, err = ss.PathValues("/v1/other/100")
	Expect(t, err, Failed())
	_, err = ss.PathValues("/v1/other")
	Expect(t, err, Failed())

	encoded := ss.Encode(path.Values{"userID": "200"})
	Expect(t, encoded, Equal("/v1/user/200"))

	encoded = ss.Encode(path.Values{})
	Expect(t, encoded, Equal("/v1/user/-"))

	ss = path.ParseSegments("/v1/{users...}/commit/{ref}")
	Expect(t, len(ss), Equal(4))
	Expect(t, ss.ParamString(), Equal("/v1/{users...}/commit/{ref}"))

	values, err = ss.PathValues("/v1/100/101/102/commit/v1")
	Expect(t, err, Succeed())
	Expect(t, values, Equal(path.Values{"users": "100/101/102", "ref": "v1"}))

	_, err = ss.PathValues("/v1/100/101/102/pull/v1")
	Expect(t, err, Failed())

	// ref: net/http/pattern.go
	// values, err = ss.PathValues("/v1/100/101/102/do/commitOther")
	// Expect(t, err, Failed())

}
