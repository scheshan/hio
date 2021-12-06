package buf

import "strconv"

func (t *Buffer) WriteByte(b byte) error {
	if err := t.checkWrite(1); err != nil {
		return err
	}

	t.writeByte(b)

	return nil
}

func (t *Buffer) WriteBool(b bool) error {
	if b {
		return t.WriteByte(1)
	} else {
		return t.WriteByte(0)
	}
}

func (t *Buffer) WriteInt8(n int8) error {
	return t.WriteByte(byte(n))
}

func (t *Buffer) WriteUInt8(n uint8) error {
	return t.WriteInt8(int8(n))
}

func (t *Buffer) WriteInt16(n int16) error {
	return t.WriteUInt16(uint16(n))
}

func (t *Buffer) WriteUInt16(n uint16) error {
	if err := t.checkWrite(2); err != nil {
		return err
	}

	t.writeUInt16(n)

	return nil
}

func (t *Buffer) WriteInt32(n int32) error {
	return t.WriteUInt32(uint32(n))
}

func (t *Buffer) WriteUInt32(n uint32) error {
	if err := t.checkWrite(4); err != nil {
		return err
	}

	t.writeUInt32(n)

	return nil
}

func (t *Buffer) WriteInt64(n int64) error {
	return t.WriteUInt64(uint64(n))
}

func (t *Buffer) WriteUInt64(n uint64) error {
	if err := t.checkWrite(8); err != nil {
		return err
	}

	t.writeUInt64(n)

	return nil
}

func (t *Buffer) WriteInt(n int) error {
	if strconv.IntSize == 32 {
		return t.WriteInt32(int32(n))
	} else {
		return t.WriteInt64(int64(n))
	}
}

func (t *Buffer) WriteUInt(n uint) error {
	if strconv.IntSize == 32 {
		return t.WriteUInt32(uint32(n))
	} else {
		return t.WriteUInt64(uint64(n))
	}
}

func (t *Buffer) WriteString(str string) error {
	return t.WriteBytes(stringToBytes(str))
}

func (t *Buffer) WriteBytes(data []byte) (err error) {
	if err = t.checkWrite(len(data)); err != nil {
		return err
	}

	t.size += len(data)

	if t.tail == nil {
		t.addNewNodeData(data)
		return
	}

	n := t.tail.writeBytes(data)
	if n == len(data) {
		return
	}

	t.addNewNodeData(data[n:])
	return
}

func (t *Buffer) Append(buf *Buffer) error {
	if err := t.checkWrite(buf.size); err != nil {
		return err
	}

	if buf.head == nil {
		return nil
	}

	t.addTail(buf.head)
	t.tail = buf.tail

	t.size += buf.size

	buf.head = nil
	buf.tail = nil
	buf.size = 0

	return nil
}
