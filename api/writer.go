package api

import "io"

// NewChunkWriter creates an io.Writer from a `ChunkSender`.
// This can be useful for passing a stream down into lower-level parts of the
// system w/o having to manually deal with the chunking required by the protocol.
func NewChunkWriter(sender ChunkSender) io.Writer {
	return &chunkWriter{sender}
}

// ChunkSender is when creating a chunk writer to be able to accept different
// streaming endpoints from the API server.
type ChunkSender interface {
	Send(*Chunk) error
}

type chunkWriter struct {
	w ChunkSender
}

func (w *chunkWriter) Write(b []byte) (int, error) {
	n := len(b)
	err := w.w.Send(&Chunk{
		Chunk: b[:n],
	})

	if err != nil {
		return 0, err
	}

	return n, nil
}
