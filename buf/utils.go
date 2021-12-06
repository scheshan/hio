package buf

import (
	"reflect"
	"unsafe"
)

func bytesToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

func stringToBytes(str string) (data []byte) {
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&str)).Data)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	hdr.Data = uintptr(p)
	hdr.Cap = len(str)
	hdr.Len = len(str)
	return data
}
