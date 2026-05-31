package transformer_test

import (
	"context"
	"io"
	"reflect"
	"testing"

	. "github.com/xoctopus/x/testx"
	"google.golang.org/protobuf/proto"

	"github.com/xoctopus/httpx/internal/payload/content"
	"github.com/xoctopus/httpx/internal/payload/transformer"
	"github.com/xoctopus/httpx/internal/payload/transformer/testdata"
	"github.com/xoctopus/httpx/internal/testutil"
)

type InBody struct {
	Name    string   `json:"name"`
	Data    []byte   `json:"data"`
	Ignored any      `json:"-"`
	List    []string `json:"list"`
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

type File struct {
	filename string
	media    string
	content  []byte
}

var (
	_ content.MediaTypeDescriber = (*File)(nil)
	_ content.FilenameModifier   = (*File)(nil)
	_ content.WithFilename       = (*File)(nil)
	_ content.Modifier           = (*File)(nil)
	_ content.ReaderFrom         = (*File)(nil)
	_ io.ReadCloser              = (*File)(nil)
)

func (f File) Filename() string {
	return f.filename
}

func (f *File) SetFilename(name string) {
	f.filename = name
}

func (f File) ContentType() string {
	return f.media
}

func (f *File) SetContentType(t string) {
	f.media = t
}

func (f *File) ReadFrom(rc io.ReadCloser) (int64, error) {
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		return 0, err
	}
	f.content = data
	return int64(len(data)), nil
}

func (f File) Read(dst []byte) (int, error) {
	return copy(dst, f.content), io.EOF
}

func (f File) Close() error {
	return nil
}

func TestTransformer(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		type T struct {
			InBody `in:"body"`
		}
		op1 := T{
			InBody: InBody{
				Name: "abc",
				Data: []byte("def"),
				List: []string{"v1", "v2"},
			},
		}

		req, err := transformer.NewRequest(context.Background(), "POST", "/root", op1)
		Expect(t, err, Succeed())
		Expect(t, req, testutil.BeRequest(`
POST /root HTTP/1.1
Content-Type: application/json; charset=utf-8

{"name":"abc","data":"ZGVm","list":["v1","v2"]}
`))

		op2 := T{}
		err = transformer.UnmarshalUnderlyingRequest(req, &op2)
		Expect(t, err, Succeed())
		Expect(t, reflect.DeepEqual(op1, op2), BeTrue())
	})

	t.Run("OCTET", func(t *testing.T) {
		type T struct {
			Body string `in:"body" mime:"octet"`
		}

		op1 := T{Body: "test"}

		req, err := transformer.NewRequest(context.Background(), "POST", "/", op1)
		Expect(t, err, Succeed())
		Expect(t, req, testutil.BeRequest(`
POST / HTTP/1.1
Content-Length: 4
Content-Type: application/octet-stream

test
`))
		op2 := T{}
		err = transformer.UnmarshalUnderlyingRequest(req, &op2)
		Expect(t, err, Succeed())
		Expect(t, reflect.DeepEqual(op1, op2), BeTrue())
	})

	t.Run("TEXT", func(t *testing.T) {
		type T struct {
			Body string `in:"body" mime:"text"`
		}

		op1 := T{Body: "test"}
		req, err := transformer.NewRequest(context.Background(), "POST", "/", op1)
		Expect(t, err, Succeed())
		Expect(t, req, testutil.BeRequest(`
POST / HTTP/1.1
Content-Length: 4
Content-Type: text/plain; charset=utf-8

test
`))
		op2 := T{}
		err = transformer.UnmarshalUnderlyingRequest(req, &op2)
		Expect(t, err, Succeed())
		Expect(t, reflect.DeepEqual(op1, op2), BeTrue())
	})

	t.Run("URLENCODED", func(t *testing.T) {
		type Data struct {
			A      string   `json:"a"`
			B      int      `json:"b"`
			C      bool     `json:"c"`
			Filter []string `json:"filter"`
		}
		type T struct {
			Data `in:"body" mime:"urlencoded"`
		}

		op1 := T{
			Data: Data{
				A:      "string",
				B:      5,
				C:      true,
				Filter: []string{"p1", "p2"},
			},
		}
		req, err := transformer.NewRequest(context.Background(), "POST", "/", op1)
		Expect(t, err, Succeed())
		Expect(t, req, testutil.BeRequest(`
POST / HTTP/1.1
Content-Type: application/x-www-form-urlencoded; param=value

a=string&b=5&c=true&filter=p1&filter=p2
`))

		op2 := T{}
		err = transformer.UnmarshalUnderlyingRequest(req, &op2)
		Expect(t, err, Succeed())
		Expect(t, reflect.DeepEqual(op1, op2), BeTrue())
	})

	t.Run("MULTIPART", func(t *testing.T) {
		type Data struct {
			Part0 string  `json:"string"`
			Part1 int     `json:"int"`
			Part2 float32 `json:"float32"`
			Part3 []bool  `json:"booleans"`
			Part4 File    `json:"file"`
			Part5 []File  `json:"files"`
		}
		type T struct {
			Data `in:"body" mime:"multipart"`
		}

		op1 := T{
			Data: Data{
				Part0: "abc",
				Part1: 1,
				Part2: 2.01,
				Part3: []bool{true, false},
				Part4: File{
					filename: "part4.txt",
					media:    "text/plain",
					content:  []byte("part4-content"),
				},
				Part5: []File{
					{
						filename: "part5-0.txt",
						media:    "text/plain",
						content:  []byte("part5-0-content"),
					},
					{
						filename: "part5-1.txt",
						media:    "text/plain",
						content:  []byte("part5-1-content"),
					},
				},
			},
		}

		req, err := transformer.NewRequest(context.Background(), "POST", "/", op1)
		Expect(t, err, Succeed())
		Expect(t, req, testutil.BeRequest(`
POST / HTTP/1.1
Content-Type: multipart/form-data; boundary=boundary1

--boundary1
Content-Disposition: form-data; name=string
Content-Length: 3
Content-Type: text/plain; charset=utf-8

abc
--boundary1
Content-Disposition: form-data; name=int
Content-Type: application/json; charset=utf-8

1
--boundary1
Content-Disposition: form-data; name=float32
Content-Type: application/json; charset=utf-8

2.01
--boundary1
Content-Disposition: form-data; name=booleans
Content-Type: application/json; charset=utf-8

true
--boundary1
Content-Disposition: form-data; name=booleans
Content-Type: application/json; charset=utf-8

false
--boundary1
Content-Disposition: form-data; filename=part4.txt; name=file
Content-Type: text/plain

part4-content
--boundary1
Content-Disposition: form-data; filename=part5-0.txt; name=files
Content-Type: text/plain

part5-0-content
--boundary1
Content-Disposition: form-data; filename=part5-1.txt; name=files
Content-Type: text/plain

part5-1-content
--boundary1--
`))

		op2 := T{}
		err = transformer.UnmarshalUnderlyingRequest(req, &op2)
		Expect(t, err, Succeed())
		Expect(t, reflect.DeepEqual(op1, op2), BeTrue())
	})

	t.Run("PROTOBUF", func(t *testing.T) {
		type T struct {
			Data testdata.UserProfile `in:"body" mime:"proto"`
		}

		op1 := T{
			Data: testdata.UserProfile{
				Id:    100,
				Name:  "saito",
				Email: "saito@xoctopus.net",
				Tags:  []string{"a", "b", "c"},
				Role:  testdata.UserRole_ROLE_ADMIN,
			},
		}
		req, err := transformer.NewRequest(context.Background(), "POST", "/", &op1)
		Expect(t, err, Succeed())

		op2 := T{}
		err = transformer.UnmarshalUnderlyingRequest(req, &op2)
		Expect(t, err, Succeed())
		Expect(t, proto.Equal(&op1.Data, &op2.Data), BeTrue())
	})
}
