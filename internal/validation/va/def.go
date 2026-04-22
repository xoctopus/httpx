package va

import (
	"math/bits"
	"strconv"
)

type Numeric interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

type Enum interface {
	Numeric | string
}

func ParseValue[T Enum](data string) (T, error) {
	var (
		x   any
		err error
	)
	switch any(*new(T)).(type) {
	case int:
		x, err = strconv.ParseInt(data, 10, bits.UintSize)
		x = int(x.(int64))
	case int8:
		x, err = strconv.ParseInt(data, 10, 8)
		x = int8(x.(int64))
	case int16:
		x, err = strconv.ParseInt(data, 10, 16)
		x = int16(x.(int64))
	case int32:
		x, err = strconv.ParseInt(data, 10, 32)
		x = int32(x.(int64))
	case int64:
		x, err = strconv.ParseInt(data, 10, 32)
	case uint:
		x, err = strconv.ParseUint(data, 10, bits.UintSize)
		x = uint(x.(uint64))
	case uint8:
		x, err = strconv.ParseUint(data, 10, 8)
		x = uint8(x.(uint64))
	case uint16:
		x, err = strconv.ParseUint(data, 10, 16)
		x = uint16(x.(uint64))
	case uint32:
		x, err = strconv.ParseUint(data, 10, 32)
		x = uint32(x.(uint64))
	case uint64:
		x, err = strconv.ParseUint(data, 10, 64)
	case float32:
		x, err = strconv.ParseFloat(data, 32)
		x = float32(x.(float64))
	case float64:
		x, err = strconv.ParseFloat(data, 64)
	default:
		x = data
	}
	if err != nil {
		x = *new(T)
	}
	return x.(T), err
}

// func MinMax[T Numeric]() (minimum, maximum T) {
// 	switch any(*new(T)).(type) {
// 	case int:
// 		return T(int(math.MinInt)), T(int(math.MaxInt))
// 	case int8:
// 		return T(int16(math.MinInt16)), T(int16(math.MaxInt16))
// 	case int16:
// 		return T(int16(math.MinInt16)), T(int16(math.MaxInt16))
// 	case int32:
// 		return T(int32(math.MinInt32)), T(int32(math.MaxInt32))
// 	case int64:
// 		return T(int64(math.MinInt64)), T(int64(math.MaxInt64))
// 	case uint:
// 		return T(uint(0)), T(uint(math.MaxUint))
// 	case uint8:
// 		return T(uint8(0)), T(uint8(math.MaxUint8))
// 	case uint16:
// 		return T(uint16(0)), T(uint16(math.MaxUint16))
// 	case uint32:
// 		return T(uint32(0)), T(uint32(math.MaxUint32))
// 	case uint64:
// 		return T(uint64(0)), T(uint64(math.MaxUint64))
// 	case float32:
// 		return T(float32(math.SmallestNonzeroFloat32)), T(float32(math.MaxFloat32))
// 	default:
// 		return T(float64(math.SmallestNonzeroFloat64)), T(float64(math.MaxFloat32))
// 	}
// }
