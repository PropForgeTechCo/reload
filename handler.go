package reload

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

type wrapper struct {
	http.ResponseWriter
	header int
	buf    *bytes.Buffer
}

func (w *wrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		Logger.Println("HTTP handler called Hijack() but the underlying responseWriter did not support it")
	}
	return hijacker.Hijack()
}

func (w *wrapper) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		Logger.Println("HTTP handler called Flush() but the underlying responseWriter did not support it")
	}
	flusher.Flush()
}

func (w *wrapper) WriteHeader(code int) {
	w.header = code
}

func findAndInsertAfter(src, match []byte, value string) []byte {
	index := bytes.Index(src, match)
	if index == -1 {
		return src
	}

	buf := &bytes.Buffer{}
	buf.Write(src[:index+len(match)])
	buf.WriteString(value)
	buf.Write(src[len(match)+index:])

	return buf.Bytes()
}

func (w *wrapper) Write(b []byte) (int, error) {
	return w.buf.Write(b)
}
