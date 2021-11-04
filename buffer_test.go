package hio

import "testing"

func TestBuffer_Write(t *testing.T) {
	buf := &Buffer{}
	buf.mp = NewMemoryPool()

	data := []byte{1, 2, 3}
	_, _ = buf.Write(data)

	data = []byte{4, 5, 6}
	buf.Write(data)
}

func TestBuffer_Read(t *testing.T) {
	buf := &Buffer{}
	buf.mp = NewMemoryPool()

	buf.Write([]byte{1, 2, 3})
	buf.Write([]byte{4, 5, 6})

	data := make([]byte, 8, 8)
	n, _ := buf.Read(data)

	if n != 6 {
		t.Fatalf("must be 6")
	}

	buf.Write([]byte{7, 8, 9, 10, 11, 12})
	data = make([]byte, 4, 4)
	n, _ = buf.Read(data)
	if n != 4 {
		t.Fatalf("must be 4")
	}

	n, _ = buf.Read(data)
	if n != 2 {
		t.Fatalf("must be 2")
	}
}
