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

func ParseEnumValue[T Enum](data string) (T, error) {
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
