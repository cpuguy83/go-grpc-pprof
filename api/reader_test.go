package api

import (
	"bytes"
	"io"
	"testing"
)

func newMockChunkRecv(chunks ...*Chunk) *mockChunkRecv {
	r := &mockChunkRecv{}
	for _, c := range chunks {
		r.chunks = append(r.chunks, c)
	}
	return r
}

type mockChunkRecv struct {
	chunks []*Chunk
}

func (m *mockChunkRecv) Recv() (*Chunk, error) {
	if len(m.chunks) == 0 {
		return nil, io.EOF
	}

	chunk := m.chunks[0]
	m.chunks = m.chunks[1:]
	return chunk, nil
}

func (m *mockChunkRecv) Size() int {
	var size int
	for _, c := range m.chunks {
		size += len(c.Chunk)
	}
	return size
}

func TestReader(t *testing.T) {
	testData := []byte("this is a test")
	testChunks := []*Chunk{
		&Chunk{Chunk: testData},
		&Chunk{Chunk: testData},
		&Chunk{Chunk: testData},
		&Chunk{Chunk: testData},
		&Chunk{Chunk: testData},
	}
	cr := newMockChunkRecv(testChunks...)
	r := NewChunkReader(cr)

	var nr int
	buf := make([]byte, 8)
	total := cr.Size()

	allData := bytes.Repeat(testData, len(testChunks))
	for nr < total {
		read, err := r.Read(buf)
		nr += read
		if err != nil {
			if nr != total {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(buf[:read], allData[nr-read:nr]) {
			t.Fatalf("%s != %s", string(buf[:read]), string(allData[nr-read:nr]))
		}
	}
}
