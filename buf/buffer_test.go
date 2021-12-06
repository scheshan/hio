package buf

import "testing"

func TestBuffer_WriteBool(t *testing.T) {
	buf := NewBuffer()

	input := true
	buf.WriteBool(input)

	b, err := buf.ReadBool()
	if err != nil {
		t.Fatal(err)
	}
	if b != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteInt16(t *testing.T) {
	buf := NewBuffer()

	var input int16 = 3000
	buf.WriteInt16(input)

	n, err := buf.ReadInt16()
	if err != nil {
		t.Fatal(err)
	}
	if n != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteInt32(t *testing.T) {
	buf := NewBuffer()

	var input int32 = 300000
	buf.WriteInt32(input)

	n, err := buf.ReadInt32()
	if err != nil {
		t.Fatal(err)
	}
	if n != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteUInt32(t *testing.T) {
	buf := NewBuffer()

	var input uint32 = 300000000
	buf.WriteUInt32(input)

	n, err := buf.ReadUInt32()
	if err != nil {
		t.Fatal(err)
	}
	if n != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteInt64(t *testing.T) {
	buf := NewBuffer()

	var input int64 = 300000
	buf.WriteInt64(input)

	n, err := buf.ReadInt64()
	if err != nil {
		t.Fatal(err)
	}
	if n != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteUInt64(t *testing.T) {
	buf := NewBuffer()

	var input uint64 = 300000000000000
	buf.WriteUInt64(input)

	n, err := buf.ReadUInt64()
	if err != nil {
		t.Fatal(err)
	}
	if n != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteBytes(t *testing.T) {
	buf := NewBuffer()

	input := []byte{1, 2, 3, 4}

	buf.WriteBytes(input)

	b, err := buf.ReadBytes(4)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(b); i++ {
		if b[i] != input[i] {
			t.Fail()
		}
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}

func TestBuffer_WriteString(t *testing.T) {
	buf := NewBuffer()

	input := "hello world"

	buf.WriteString(input)

	str, err := buf.ReadString(11)
	if err != nil {
		t.Fatal(err)
	}
	if str != input {
		t.Fail()
	}
	if buf.ReadableBytes() > 0 {
		t.Fail()
	}
}
