package h1

import (
	"io"
	"strconv"
)

type Response struct {
	upstream io.Writer

	buf []byte // Note: Do not use append() to add bytes to this buffer. Use Write() instead. This is to avoid unnecessary memory allocations.
	n   int

	// Itoa Buffer
	itoaBuf []byte // buffer for itoa
}

func (r *Response) Reset() {
	r.n = 0
	r.buf = r.buf[:0]
}

func (r *Response) Flush() error {
	if r.upstream == nil {
		return nil
	}

	_, err := r.upstream.Write(r.buf)
	if err != nil {
		return err
	}

	r.Reset()
	return nil
}

func (r *Response) Write(b []byte) (int, error) {
	n := copy(r.buf[r.n:], b) // copy to buffer
	r.n += n
	if len(r.buf) < cap(r.buf) {
		return n, nil
	}

	// buffer is full, flush it
	err := r.Flush()
	if err != nil {
		return 0, err
	}

	// If b is bigger than buffer, write it directly
	if len(b)-n > cap(r.buf) {
		_, err = r.upstream.Write(b[n:])
		return len(b), err
	}

	// copy b to buffer
	copy(r.buf, b[n:])
	r.n = len(b) - n
	return n, nil
}

func (r *Response) WriteString(s string) (int, error) {
	return r.Write(stringToBytes(s))
}

func (r *Response) WriteInt(i int) (int, error) {
	r.itoaBuf = r.itoaBuf[:0]
	r.itoaBuf = strconv.AppendInt(r.itoaBuf, int64(i), 10)
	return r.Write(r.itoaBuf)
}

func (r *Response) WriteUint(u uint) (int, error) {
	r.itoaBuf = r.itoaBuf[:0]
	r.itoaBuf = strconv.AppendUint(r.itoaBuf, uint64(u), 10)
	return r.Write(r.itoaBuf)
}

func (r *Response) WriteStatusLine(status int) error {
	_, err := r.Write(GetStatusLine(status))
	return err
}
