package hio

import "testing"

func TestBuffer_Write(t *testing.T) {
	buf := &Buffer{}

	buf.Write([]byte{1, 2, 3})
	if buf.ReadableBytes() != 3 {
		t.Fail()
	}

	buf.Write([]byte{4, 5, 6, 7, 8, 9})
	if buf.ReadableBytes() != 9 {
		t.Fail()
	}
}

func TestBuffer_ReadByte(t *testing.T) {
	buf := &Buffer{}
	buf.Write([]byte{1, 2, 3})

	b, _ := buf.ReadByte()
	if b != 1 {
		t.Fail()
	}
	b, _ = buf.ReadByte()
	if b != 2 {
		t.Fail()
	}
	b, _ = buf.ReadByte()
	if b != 3 {
		t.Fail()
	}
}

func TestBuffer_ReadBytes(t *testing.T) {
	buf := &Buffer{}
	buf.Write([]byte{1, 2, 3})

	b, err := buf.ReadBytes(3)
	if err != nil {
		t.Fail()
	}

	if b.Len() != 3 {
		t.Fail()
	}
}
