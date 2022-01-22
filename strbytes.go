//go:build !appengine && !nounsafe

package h1

import (
	"reflect"
	"unsafe"
)

func stringToBytes(s string) []byte {
	//#nosec
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	//#nosec
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: strHeader.Data,
		Len:  strHeader.Len,
		Cap:  strHeader.Len,
	}))
}

func bytesToString(b []byte) string {
	//#nosec
	return *(*string)(unsafe.Pointer(&b))
}
