package va_test

import (
	"testing"

	. "github.com/xoctopus/x/testx"

	"github.com/xoctopus/httpx/internal/validation/va"
)

func TestParseValue(t *testing.T) {
	i, err := va.ParseValue[int]("1")
	Expect(t, i, Equal(1))

	i8, err := va.ParseValue[int8]("128")
	Expect(t, err, Failed())
	Expect(t, i8, Equal(int8(0)))

	i16, err := va.ParseValue[int16]("128")
	Expect(t, i16, Equal(int16(128)))

	i32, err := va.ParseValue[int32]("128")
	Expect(t, i32, Equal(int32(128)))

	i64, err := va.ParseValue[int64]("128")
	Expect(t, i64, Equal(int64(128)))

	u, err := va.ParseValue[uint]("1")
	Expect(t, u, Equal(uint(1)))

	u8, err := va.ParseValue[uint8]("257")
	Expect(t, err, Failed())
	Expect(t, u8, Equal(uint8(0)))

	u16, err := va.ParseValue[uint16]("128")
	Expect(t, u16, Equal(uint16(128)))

	u32, err := va.ParseValue[uint32]("128")
	Expect(t, u32, Equal(uint32(128)))

	f32, err := va.ParseValue[float32]("128.001")
	Expect(t, f32, Equal(float32(128.001)))

	f64, err := va.ParseValue[float64]("128.001")
	Expect(t, f64, Equal(128.001))

	u64, err := va.ParseValue[uint64]("128")
	Expect(t, u64, Equal(uint64(128)))

	str, err := va.ParseValue[string]("128")
	Expect(t, str, Equal("128"))
}
