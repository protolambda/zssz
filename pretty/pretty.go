package pretty

import (
	"io"
	"strings"
	"unsafe"
)

type PrettyFn func(indent uint32, w *PrettyWriter, p unsafe.Pointer)

type PrettyWriter struct {
	w      io.Writer
	indent string
}

func NewPrettyWriter(w io.Writer, indent string) *PrettyWriter {
	return &PrettyWriter{w: w, indent: indent}
}

// Write writes len(p) bytes from p to the underlying accumulated buffer.
func (pw *PrettyWriter) Write(p string) {
	_, _ = pw.w.Write([]byte(p))
}

// Write writes len(p) bytes from p to the underlying accumulated buffer.
func (pw *PrettyWriter) WriteIndent(indent uint32) {
	pw.Write(strings.Repeat(pw.indent, int(indent)))
}
