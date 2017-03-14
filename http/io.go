package http

import "io"

type writeFlusher struct {
	w io.Writer
	f flusher
}
type flusher interface {
	Flush()
}

func makeFlusher(w io.Writer) io.Writer {
	if f, ok := w.(flusher); ok {
		f.Flush()
		return &writeFlusher{w, f}
	}
	return w
}

func (w *writeFlusher) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.f.Flush()
	return n, err
}
