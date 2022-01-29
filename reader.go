package h1

import (
	"io"
	"sync"
)

type RequestReader struct {
	R io.Reader

	ReadBuffer []byte
	NextBuffer []byte

	Request Request
}

func (r *RequestReader) Reset() {
	r.ReadBuffer = r.ReadBuffer[:cap(r.ReadBuffer)]
	r.NextBuffer = r.ReadBuffer[:0]
	r.Request.Reset()
}

func (r *RequestReader) Next() (remainint int, err error) {
	var retryCount int = 0

	if r.Remaining() == 0 {
		n, err := r.R.Read(r.ReadBuffer[:cap(r.ReadBuffer)])
		if err != nil {
			return 0, err
		}
		r.NextBuffer = r.ReadBuffer[:n]
	}

parse:
	// Reset the request
	r.Request.Reset()

	// Read request line
	r.NextBuffer, err = ParseRequestLine(&r.Request, r.NextBuffer)
	if err != nil {
		if err == ErrBufferTooSmall {
			// Buffer is too small, read more bytes

			// Copy the remaining bytes to the read buffer
			n0 := copy(r.ReadBuffer[:cap(r.ReadBuffer)], r.NextBuffer)

			// Read more bytes
			n1, err := r.R.Read(r.ReadBuffer[n0:cap(r.ReadBuffer)])
			if err != nil {
				return 0, err
			}

			// Set the next buffer to the read buffer
			r.NextBuffer = r.ReadBuffer[:n0+n1]

			// Retry parsing
			retryCount++
			if retryCount > 2 {
				return 0, ErrBufferTooSmall
			}
			goto parse
		}
	}

	// Read headers
	r.NextBuffer, err = ParseHeaders(&r.Request, r.NextBuffer)
	if err != nil {
		if err == ErrBufferTooSmall {
			// Buffer is too small, read more bytes

			// Copy the remaining bytes to the read buffer
			n0 := copy(r.ReadBuffer, r.NextBuffer)

			// Read more bytes
			n1, err := r.R.Read(r.ReadBuffer[n0:cap(r.ReadBuffer)])
			if err != nil {
				return 0, err
			}

			// Set the next buffer to the read buffer
			r.NextBuffer = r.ReadBuffer[:n0+n1]

			// Retry parsing
			retryCount++
			if retryCount > 2 {
				return 0, ErrBufferTooSmall
			}
			goto parse
		}
	}

	return len(r.NextBuffer), nil
}

func (r *RequestReader) Remaining() int {
	return len(r.NextBuffer)
}

type BodyReader struct {
	Upstream *RequestReader

	Limit int
	Index int
}

func (r *BodyReader) reset() {
	r.Upstream = nil

	r.Limit = 0
	r.Index = 0
}

var BodyReaderPool = &sync.Pool{
	New: func() any {
		return &BodyReader{}
	},
}

func GetBodyReader() *BodyReader {
	return BodyReaderPool.Get().(*BodyReader)
}

func PutBodyReader(r *BodyReader) {
	r.reset()
	BodyReaderPool.Put(r)
}
