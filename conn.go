package hio

import (
	"sync"
	"syscall"
)

type Conn struct {
	id        uint64
	sa        syscall.Sockaddr
	fd        int
	in        *Buffer
	out       *Buffer
	loop      *EventLoop
	writeMask bool
	mutex     *sync.Mutex
}

func (t *Conn) Close() error {
	return nil
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
}

//#region reader logic

func (t *Conn) ReadableBytes() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadableBytes()
}

func (t *Conn) CanRead() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.CanRead()
}

func (t *Conn) ReadByte() (byte, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadByte()
}

func (t *Conn) ReadBool() (bool, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadBool()
}

func (t *Conn) ReadInt8() (int8, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadInt8()
}

func (t *Conn) ReadUInt8() (uint8, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadUInt8()
}

func (t *Conn) ReadInt16() (int16, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadInt16()
}

func (t *Conn) ReadUInt16() (uint16, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadUInt16()
}

func (t *Conn) ReadInt32() (int32, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadInt32()
}

func (t *Conn) ReadUInt32() (uint32, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadUInt32()
}

func (t *Conn) ReadInt64() (int64, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadInt64()
}

func (t *Conn) ReadUInt64() (uint64, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadUInt64()
}

func (t *Conn) ReadInt() (int, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadInt()
}

func (t *Conn) ReadUInt() (uint, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadUInt()
}

func (t *Conn) ReadBytes(n int) ([]byte, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadBytes(n)
}

func (t *Conn) ReadString(n int) (string, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.in.ReadString(n)
}

//#endregion

//#region writer logic

func (t *Conn) WriteByte(b byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteByte(b)

	t.setWriteMask()
}

func (t *Conn) WriteBool(b bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteBool(b)

	t.setWriteMask()
}

func (t *Conn) WriteInt8(n int8) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteInt8(n)

	t.setWriteMask()
}

func (t *Conn) WriteUInt8(n uint8) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteUInt8(n)

	t.setWriteMask()
}

func (t *Conn) WriteInt16(n int16) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteInt16(n)

	t.setWriteMask()
}

func (t *Conn) WriteUInt16(n uint16) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteUInt16(n)

	t.setWriteMask()
}

func (t *Conn) WriteInt32(n int32) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteInt32(n)

	t.setWriteMask()
}

func (t *Conn) WriteUInt32(n uint32) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteUInt32(n)

	t.setWriteMask()
}

func (t *Conn) WriteInt64(n int64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteInt64(n)

	t.setWriteMask()
}

func (t *Conn) WriteUInt64(n uint64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteUInt64(n)

	t.setWriteMask()
}

func (t *Conn) WriteInt(n int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteInt(n)

	t.setWriteMask()
}

func (t *Conn) WriteUInt(n uint) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteUInt(n)

	t.setWriteMask()
}

func (t *Conn) WriteString(str string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteString(str)

	t.setWriteMask()
}

func (t *Conn) WriteBytes(data []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.WriteBytes(data)

	t.setWriteMask()
}

//#endregion

func (t *Conn) setWriteMask() {
	if t.writeMask {
		return
	}
	t.writeMask = true
}

func (t *Conn) doClose() {
	syscall.Close(t.fd)
	t.in.Release()
	t.out.Release()
}

func newConn(id uint64, sa syscall.Sockaddr, fd int) *Conn {
	conn := &Conn{
		id:    id,
		sa:    sa,
		fd:    fd,
		in:    &Buffer{},
		out:   &Buffer{},
		mutex: &sync.Mutex{},
	}

	return conn
}
