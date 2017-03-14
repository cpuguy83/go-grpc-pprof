package api

import (
	"bytes"
	"io"
	"testing"
)

type mockChunkSender struct {
	chunks []*Chunk
}

func (m *mockChunkSender) Send(c *Chunk) error {
	m.chunks = append(m.chunks, c)
	return nil
}

func (m *mockChunkSender) Bytes() []byte {
	var out []byte
	for _, c := range m.chunks {
		out = append(out, c.Chunk...)
	}
	return out
}

func TestChunkWriter(t *testing.T) {
	m := &mockChunkSender{}
	w := NewChunkWriter(m)

	testData := []byte("this is a test")
	testChunks := []*Chunk{
		{Chunk: testData},
		{Chunk: testData},
		{Chunk: testData},
		{Chunk: testData},
		{Chunk: testData},
	}
	r := NewChunkReader(newMockChunkRecv(testChunks...))

	nw, err := io.Copy(w, r)
	if err != nil {
		t.Fatal(err)
	}

	allTestBytes := bytes.Repeat(testData, len(testChunks))
	total := len(allTestBytes)
	if nw != int64(total) {
		t.Fatalf("expected %d wrriten, got %d", total, nw)
	}

	allBytes := m.Bytes()
	if !bytes.Equal(allBytes, allTestBytes) {
		t.Fatal("bytes sent != bytes received")
	}
}
