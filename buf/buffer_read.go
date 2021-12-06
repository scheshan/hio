package buf

import (
	"golang.org/x/sys/unix"
	"strconv"
)

//ReadableBytes return the total bytes can be read in the buffer
func (t *Buffer) ReadableBytes() int {
	return t.size
}

//ReadByte read a single byte. If there is no enough data, return error
func (t *Buffer) ReadByte() (b byte, err error) {
	if err = t.checkSize(1); err != nil {
		return 0, err
	}

	for {
		if t.head.readableBytes() > 0 {
			b = t.readHeadBytes(1)[0]
			break
		}

		t.head = t.head.next
	}

	return
}

//ReadBool return a bool value. If there is no enough data, return error
func (t *Buffer) ReadBool() (bool, error) {
	b, err := t.ReadByte()
	if err != nil {
		return false, err
	}

	return b != 0, nil
}

//ReadInt8 return a int8 value. If there is no enough data, return error
func (t *Buffer) ReadInt8() (int8, error) {
	b, err := t.ReadByte()
	if err != nil {
		return 0, err
	}

	return int8(b), nil
}

//ReadUInt8 return a uint8 value. If there is no enough data, return error
func (t *Buffer) ReadUInt8() (uint8, error) {
	return t.ReadByte()
}

//ReadInt16 return a int16 value. If there is no enough data, return error
func (t *Buffer) ReadInt16() (int16, error) {
	res, err := t.ReadUInt16()
	if err != nil {
		return 0, err
	}

	return int16(res), nil
}

//ReadUInt16 return a uint16 value. If there is no enough data, return error
func (t *Buffer) ReadUInt16() (uint16, error) {
	if err := t.checkSize(2); err != nil {
		return 0, err
	}

	var res uint16
	if t.head.readableBytes() >= 2 {
		b := t.readHeadBytes(2)
		res = uint16(b[0])<<8 | uint16(b[1])
	} else {
		b1, _ := t.ReadByte()
		b2, _ := t.ReadByte()
		res = (uint16(b1) << 8) | uint16(b2)
	}

	return res, nil
}

//ReadInt32 return a uint32 value. If there is no enough data, return error
func (t *Buffer) ReadInt32() (int32, error) {
	n, err := t.ReadUInt32()
	if err != nil {
		return 0, err
	}

	return int32(n), nil
}

//ReadUInt32 return a uint32 value. If there is no enough data, return error
func (t *Buffer) ReadUInt32() (uint32, error) {
	if err := t.checkSize(4); err != nil {
		return 0, err
	}

	var res uint32

	if t.head.readableBytes() >= 4 {
		b := t.readHeadBytes(4)

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

//ReadInt64 return a int64 value. If there is no enough data, return error
func (t *Buffer) ReadInt64() (int64, error) {
	n, err := t.ReadUInt64()
	if err != nil {
		return 0, err
	}

	return int64(n), nil
}

//ReadUInt64 return a uint64 value. If there is no enough data, return error
func (t *Buffer) ReadUInt64() (uint64, error) {
	if err := t.checkSize(8); err != nil {
		return 0, err
	}

	var res uint64
	if t.head.readableBytes() >= 8 {
		b := t.readHeadBytes(8)

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

//ReadInt return a int value. If there is no enough data, return error
func (t *Buffer) ReadInt() (int, error) {
	if strconv.IntSize == 32 {
		n, err := t.ReadInt32()
		return int(n), err
	} else {
		n, err := t.ReadInt64()
		return int(n), err
	}
}

//ReadUInt return a uint value. If there is no enough data, return error
func (t *Buffer) ReadUInt() (uint, error) {
	if strconv.IntSize == 32 {
		n, err := t.ReadUInt32()
		return uint(n), err
	} else {
		n, err := t.ReadUInt64()
		return uint(n), err
	}
}

//ReadBytes read n bytes. If there is no enough data, return error
func (t *Buffer) ReadBytes(n int) ([]byte, error) {
	if err := t.checkSize(n); err != nil {
		return nil, err
	}

	res := make([]byte, n)

	cnt := 0
	for cnt < n {
		avail := t.head.w - t.head.r

		if avail > 0 {
			rn := n - cnt
			if rn > avail {
				rn = avail
			}

			copy(res[cnt:], t.readHeadBytes(rn))

			cnt += rn
		}
	}

	t.size -= n

	return res, nil
}

//ReadString read n bytes and convert to string. If there is no enough data, return error
func (t *Buffer) ReadString(n int) (string, error) {
	data, err := t.ReadBytes(n)
	if err != nil {
		return "", err
	}

	return bytesToString(data), nil
}

func (t *Buffer) ReadToFile(fd int) (int, error) {
	if t.ReadableBytes() == 0 {
		return 0, ErrBufferNoEnoughData
	}

	h := t.head
	n, err := unix.Write(fd, h.b[h.r:h.w])

	if err != nil {
		return 0, nil
	}

	if n > 0 {
		t.skipHeadBytes(n)
	}

	return n, nil
}
