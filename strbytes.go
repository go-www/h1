package h1

import (
	"reflect"
	"unsafe"
)

func stringToBytes(s string) []byte {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))

	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}))
}

func bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
