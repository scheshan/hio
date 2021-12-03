package hio

import (
	"errors"
	"reflect"
	"strconv"
	"unsafe"
)

var ErrBufferNoEnoughData = errors.New("buffer has no enough data to read")

type Buffer struct {
	head     *bufferNode
	tail     *bufferNode
	size     int
	readonly bool
}

//#region write logic

func (t *Buffer) WriteByte(b byte) {
	if !t.CanWrite() {
		return
	}

	t.size++

	if t.tail == nil || t.tail.writableBytes() == 0 {
		t.addNewNode(nil)
	}
	t.tail.writeByte(b)
}

func (t *Buffer) WriteBool(b bool) {
	if b {
		t.WriteByte(1)
	} else {
		t.WriteByte(0)
	}
}

func (t *Buffer) WriteInt8(n int8) {
	t.WriteByte(byte(n))
}

func (t *Buffer) WriteUInt8(n uint8) {
	t.WriteInt8(int8(n))
}

func (t *Buffer) WriteInt16(n int16) {
	if !t.CanWrite() {
		return
	}

	if t.tail != nil && t.tail.writableBytes() >= 2 {
		t.size += 2
		t.tail.writeByte(byte(n>>8), byte(n))
	} else {
		t.WriteByte(byte(n >> 8))
		t.WriteByte(byte(n))
	}
}

func (t *Buffer) WriteUInt16(n uint16) {
	t.WriteInt16(int16(n))
}

func (t *Buffer) WriteInt32(n int32) {
	if !t.CanWrite() {
		return
	}

	if t.tail != nil && t.tail.writableBytes() >= 4 {
		t.size += 4
		t.tail.writeByte(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	} else {
		t.WriteInt16(int16(n >> 16))
		t.WriteInt16(int16(n))
	}
}

func (t *Buffer) WriteUInt32(n uint32) {
	t.WriteInt32(int32(n))
}

func (t *Buffer) WriteInt64(n int64) {
	if !t.CanWrite() {
		return
	}

	if t.tail != nil && t.tail.writableBytes() >= 8 {
		t.size += 8
		t.tail.writeByte(byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
	} else {
		t.WriteInt32(int32(n >> 32))
		t.WriteInt32(int32(n))
	}
}

func (t *Buffer) WriteUInt64(n uint64) {
	t.WriteInt64(int64(n))
}

func (t *Buffer) WriteInt(n int) {
	if strconv.IntSize == 32 {
		t.WriteInt32(int32(n))
	} else {
		t.WriteInt64(int64(n))
	}
}

func (t *Buffer) WriteUInt(n uint) {
	if strconv.IntSize == 32 {
		t.WriteUInt32(uint32(n))
	} else {
		t.WriteUInt64(uint64(n))
	}
}

func (t *Buffer) WriteString(str string) {
	t.WriteBytes(t.stringToBytes(str))
}

func (t *Buffer) WriteBytes(data []byte) {
	if !t.CanWrite() {
		return
	}

	t.size += len(data)

	if t.tail == nil {
		t.addNewNode(data)
		return
	}

	n := t.tail.writeBytes(data)
	if n == len(data) {
		return
	}

	t.addNewNode(data[n:])
}

//#endregion

//#region read logic

func (t *Buffer) ReadableBytes() int {
	return t.size
}

func (t *Buffer) ReadByte() (byte, error) {
	if err := t.checkSize(1); err != nil {
		return 0, err
	}
	t.size--

	var res byte
	for {
		if t.head.readableBytes() > 0 {
			res = t.head.readBytes(1)[0]
			t.skipNode()
			break
		}
		t.skipNode()
	}

	return res, nil
}

func (t *Buffer) ReadBool() (bool, error) {
	b, err := t.ReadByte()
	if err != nil {
		return false, err
	}

	return b != 0, nil
}

func (t *Buffer) ReadInt8() (int8, error) {
	b, err := t.ReadByte()
	if err != nil {
		return 0, err
	}

	return int8(b), nil
}

func (t *Buffer) ReadUInt8() (uint8, error) {
	return t.ReadByte()
}

func (t *Buffer) ReadInt16() (int16, error) {
	res, err := t.ReadUInt16()
	if err != nil {
		return 0, err
	}

	return int16(res), nil
}

func (t *Buffer) ReadUInt16() (uint16, error) {
	if err := t.checkSize(2); err != nil {
		return 0, err
	}

	var res uint16
	if t.head.readableBytes() >= 2 {
		b := t.head.readBytes(2)
		t.size -= 2
		t.skipNode()

		res = uint16(b[0])<<8 | uint16(b[1])
	} else {
		b1, _ := t.ReadByte()
		b2, _ := t.ReadByte()
		res = (uint16(b1) << 8) | uint16(b2)
	}

	return res, nil
}

func (t *Buffer) ReadInt32() (int32, error) {
	n, err := t.ReadUInt32()
	if err != nil {
		return 0, err
	}

	return int32(n), nil
}

func (t *Buffer) ReadUInt32() (uint32, error) {
	if err := t.checkSize(4); err != nil {
		return 0, err
	}

	var res uint32

	if t.head.readableBytes() >= 4 {
		b := t.head.readBytes(4)
		t.size -= 4
		t.skipNode()

		res |= uint32(b[0]) << 24
		res |= uint32(b[1]) << 16
		res |= uint32(b[2]) << 8
		res |= uint32(b[3])
	} else {
		n1, _ := t.ReadUInt16()
		n2, _ := t.ReadUInt16()

		res = uint32(n1)<<16 | uint32(n2)
	}

	return res, nil
}

func (t *Buffer) ReadInt64() (int64, error) {
	n, err := t.ReadUInt64()
	if err != nil {
		return 0, err
	}

	return int64(n), nil
}

func (t *Buffer) ReadUInt64() (uint64, error) {
	if err := t.checkSize(8); err != nil {
		return 0, err
	}

	var res uint64
	if t.head.readableBytes() >= 8 {
		b := t.head.readBytes(8)
		t.size -= 8
		t.skipNode()

		res |= uint64(b[0]) << 56
		res |= uint64(b[1]) << 48
		res |= uint64(b[2]) << 40
		res |= uint64(b[3]) << 32
		res |= uint64(b[4]) << 24
		res |= uint64(b[5]) << 16
		res |= uint64(b[6]) << 8
		res |= uint64(b[7])
	} else {
		n1, _ := t.ReadUInt32()
		n2, _ := t.ReadUInt32()

		res = uint64(n1)<<32 | uint64(n2)
	}

	return res, nil
}

func (t *Buffer) ReadInt() (int, error) {
	if strconv.IntSize == 32 {
		n, err := t.ReadInt32()
		return int(n), err
	} else {
		n, err := t.ReadInt64()
		return int(n), err
	}
}

func (t *Buffer) ReadUInt() (uint, error) {
	if strconv.IntSize == 32 {
		n, err := t.ReadUInt32()
		return uint(n), err
	} else {
		n, err := t.ReadUInt64()
		return uint(n), err
	}
}

func (t *Buffer) ReadBytes(n int) ([]byte, error) {
	if err := t.checkSize(n); err != nil {
		return nil, err
	}

	res := make([]byte, n)

	cnt := 0
	for cnt < n {
		if t.head.readableBytes() > 0 {
			rn := n - cnt
			if rn > t.head.readableBytes() {
				rn = t.head.readableBytes()
			}

			b := t.head.readBytes(rn)
			copy(res[cnt:], b)

			cnt += rn
		}
		t.skipNode()
	}

	t.size -= n

	return res, nil
}

func (t *Buffer) ReadString(n int) (string, error) {
	data, err := t.ReadBytes(n)
	if err != nil {
		return "", err
	}

	return t.bytesToString(data), nil
}

//#endregion

func (t *Buffer) Slice(n int) (*Buffer, error) {
	if err := t.checkSize(n); err != nil {
		return nil, err
	}

	buf := pool.getBuffer()
	buf.readonly = true
	buf.size = t.size

	cnt := 0
	for cnt < n {
		r := t.head
		r.ref++

		node := pool.getNode()
		node.r = r.r
		node.w = r.w
		node.origin = r
		node.b = r.b

		if cnt+node.readableBytes() > n {
			node.w = node.r + n - cnt
		}
		cnt += node.readableBytes()

		buf.addNodeToTail(node)
	}

	return buf, nil
}

func (t *Buffer) Release() {
	for t.head != nil {
		n := t.head.next
		t.head.next = nil
		t.head.release()

		t.head = n
	}
}

func (t *Buffer) CanRead() bool {
	return t.ReadableBytes() > 0
}

func (t *Buffer) CanWrite() bool {
	return !t.readonly
}

func (t *Buffer) addNewNode(data []byte) {
	size := 1
	if data != nil && len(data) > size {
		size = len(data)
	}

	node := pool.getNodeSize(size)

	if data != nil {
		node.writeBytes(data)
	}

	t.addNodeToTail(node)
}

func (t *Buffer) addNodeToTail(node *bufferNode) {
	if t.tail == nil {
		t.head = node
		t.tail = node
	} else {
		t.tail.next = node
		t.tail = node
	}
}

func (t *Buffer) checkSize(size int) error {
	if t.ReadableBytes() < size {
		return ErrBufferNoEnoughData
	}

	return nil
}

func (t *Buffer) bytesToString(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}

func (t *Buffer) stringToBytes(str string) (data []byte) {
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&str)).Data)
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	hdr.Data = uintptr(p)
	hdr.Cap = len(str)
	hdr.Len = len(str)
	return data
}

func (t *Buffer) skipNode() {
	if t.head.readableBytes() == 0 {
		next := t.head.next
		t.head.release()
		t.head = next
	}
}
