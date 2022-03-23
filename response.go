package h1

import (
	"io"
	"strconv"
	"sync"
)

var ResponsePool = sync.Pool{
	New: func() any {
		return &Response{
			upstream:      nil,
			buf:           make([]byte, 8192),
			itoaBuf:       make([]byte, 0, 32),
			n:             0,
			ContentLength: -1,
			//Connection:    ConnectionKeepAlive,
		}
	},
}

func GetResponse(upstream io.Writer) *Response {
	r := ResponsePool.Get().(*Response)
	r.upstream = upstream
	return r
}

func PutResponse(r *Response) {
	r.Reset()
	r.upstream = nil
	ResponsePool.Put(r)
}

type Response struct {
	upstream io.Writer

	buf []byte // Note: Do not use append() to add bytes to this buffer. Use Write() instead. This is to avoid unnecessary memory allocations.
	n   int
	err error

	// Itoa Buffer
	itoaBuf []byte // buffer for itoa

	// Standard Hop-by-Hop response headers.
	ContentLength int
	//Connection    Connection
}

func (r *Response) Reset() {
	r.n = 0
	r.ContentLength = -1
}

var DefaultFastDateServer = NewFastDateServer("h1")

var _ = func() int {
	go DefaultFastDateServer.Start()
	return 0
}()

var DateServerHeaderFunc = func() []byte {
	return DefaultFastDateServer.GetDate()
}

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the GOLICENSE file.

// Package bufio implements buffered I/O. It wraps an io.Reader or io.Writer
// object, creating another object (Reader or Writer) that also implements
// the interface but provides buffering and some help for textual I/O.

// Flush writes any buffered data to the underlying upstream.
func (r *Response) Flush() error {
	if r.err != nil {
		return r.err
	}
	if r.n == 0 {
		return nil
	}
	n, err := r.upstream.Write(r.buf[0:r.n])
	if n < r.n && err == nil {
		err = io.ErrShortWrite
	}
	if err != nil {
		if n > 0 && n < r.n {
			copy(r.buf[0:r.n-n], r.buf[n:r.n])
		}
		r.n -= n
		r.err = err
		return err
	}
	r.n = 0
	return nil
}

// Available returns how many bytes are unused in the buffer.
func (r *Response) Available() int {
	return len(r.buf) - r.n
}

// AvailableBuffer returns an empty buffer with b.Available() capacity.
// This buffer is intended to be appended to and
// passed to an immediately succeeding Write call.
// The buffer is only valid until the next write operation on b.
func (r *Response) AvailableBuffer() []byte {
	return r.buf[r.n:][:0]
}

// Buffered returns the number of bytes that have been written into the current buffer.
func (r *Response) Buffered() int {
	return r.n
}

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (r *Response) Write(p []byte) (nn int, err error) {
	for len(p) > r.Available() && r.err == nil {
		var n int
		if r.Buffered() == 0 {
			// Large write, empty buffer.
			// Write directly from p to avoid copy.
			n, r.err = r.upstream.Write(p)
		} else {
			n = copy(r.buf[r.n:], p)
			r.n += n
			r.Flush()
		}
		nn += n
		p = p[n:]
	}
	if r.err != nil {
		return nn, r.err
	}
	n := copy(r.buf[r.n:], p)
	r.n += n
	nn += n
	return nn, nil
}

// WriteString writes a string.
// It returns the number of bytes written.
// If the count is less than len(s), it also returns an error explaining
// why the write is short.
func (r *Response) WriteString(s string) (int, error) {
	nn := 0
	for len(s) > r.Available() && r.err == nil {
		n := copy(r.buf[r.n:], s)
		r.n += n
		nn += n
		s = s[n:]
		r.Flush()
	}
	if r.err != nil {
		return nn, r.err
	}
	n := copy(r.buf[r.n:], s)
	r.n += n
	nn += n
	return nn, nil
}

// WriteByte writes a single byte.
func (r *Response) WriteByte(c byte) error {
	if r.err != nil {
		return r.err
	}
	if r.Available() <= 0 && r.Flush() != nil {
		return r.err
	}
	r.buf[r.n] = c
	r.n++
	return nil
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

func (r *Response) WriteInt64(i int64) (int, error) {
	r.itoaBuf = r.itoaBuf[:0]
	r.itoaBuf = strconv.AppendInt(r.itoaBuf, i, 10)
	return r.Write(r.itoaBuf)
}

func (r *Response) WriteUint64(u uint64) (int, error) {
	r.itoaBuf = r.itoaBuf[:0]
	r.itoaBuf = strconv.AppendUint(r.itoaBuf, u, 10)
	return r.Write(r.itoaBuf)
}

func (r *Response) WriteUint64Hex(u uint64) (int, error) {
	r.itoaBuf = r.itoaBuf[:0]
	r.itoaBuf = strconv.AppendUint(r.itoaBuf, u, 16)
	return r.Write(r.itoaBuf)
}

func (r *Response) WriteStatusLine(status int) error {
	_, err := r.Write(GetStatusLine(status))
	return err
}

var contentLengthHeader = []byte("Content-Length: ")
var crlf = []byte("\r\n")

func (r *Response) WriteHeader(status int) error {
	err := r.WriteStatusLine(status)
	if err != nil {
		return err
	}
	// Write standard hop-by-hop response headers

	_, err = r.Write(DateServerHeaderFunc())
	if err != nil {
		return err
	}

	// Content-Length
	if r.ContentLength >= 0 {
		_, err = r.Write(contentLengthHeader)
		if err != nil {
			return err
		}
		_, err = r.WriteInt(r.ContentLength)
		if err != nil {
			return err
		}
		_, err = r.Write(crlf)
		if err != nil {
			return err
		}
	}

	return nil
}
