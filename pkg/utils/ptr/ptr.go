/*
Package ptr provides utility functions for converting non-addressable primitive types to pointers.
Its useful in contexts where a variable gives nil primitive type pointers semantics
(often meaning "not set") which can make it annoying to set the value.

Example

	type Foo struct {
		A *int
	}

	func main() {
		foo := Foo{
			A: ptr.Int(1)
		}
	}
*/
package ptr

func Int(v int) *int {
	return &v
}

func Int8(v int8) *int8 {
	return &v
}

func Int16(v int16) *int16 {
	return &v
}

func Int32(v int32) *int32 {
	return &v
}

func Int64(v int64) *int64 {
	return &v
}

func Uint(v uint) *uint {
	return &v
}

func Uint8(v uint8) *uint8 {
	return &v
}

func Uint16(v uint16) *uint16 {
	return &v
}

func Uint32(v uint32) *uint32 {
	return &v
}

func Uint64(v uint64) *uint64 {
	return &v
}

func Float32(v float32) *float32 {
	return &v
}

func Float64(v float64) *float64 {
	return &v
}

func String(v string) *string {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

func Byte(v byte) *byte {
	return &v
}

func Rune(v rune) *rune {
	return &v
}

func Complex64(v complex64) *complex64 {
	return &v
}

func Complex128(v complex128) *complex128 {
	return &v
}
