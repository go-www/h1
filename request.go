package h1

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"sync"
)

type Request struct {
	noCopy

	// Request line
	Method  Method
	URI     []byte
	Version []byte

	// Headers
	Headers    *Header
	lastHeader *Header

	ContentLength int64

	// Body
	Buffer *[]byte
	next   []byte
}

var requestPool = sync.Pool{
	New: func() interface{} {
		return &Request{}
	},
}

func (r *Request) Reset() {
	r.Method = MethodInvalid
	r.URI = nil
	r.Version = nil
	ReturnAllHeaders(r.Headers)
	r.Headers = nil
	r.lastHeader = nil
	r.ContentLength = 0
}

func (r *Request) GetHeader(name []byte) (*Header, bool) {
	h := r.Headers
	for h != nil {
		if stricmp(h.Name, name) {
			return h, true
		}
		h = h.nextHeader
	}
	return nil, false
}

func GetRequest() *Request {
	return requestPool.Get().(*Request)
}

func PutRequest(r *Request) {
	r.Reset()
	requestPool.Put(r)
}

var ErrInvalidMethod = errors.New("invalid method")
var ErrInvalidURI = errors.New("invalid uri")
var ErrInvalidVersion = errors.New("invalid version")

var ErrBufferTooSmall = errors.New("buffer too small")

func splitLine(src []byte) (line, rest []byte, err error) {
	idx := bytes.IndexByte(src, '\n')
	if idx < 1 { // 0: cr 1: lf
		return nil, src, ErrBufferTooSmall
	}

	if src[idx-1] == '\r' {
		line = src[:idx-1]
		rest = src[idx+1:]
		return
	}
	return src[:idx], src[idx+1:], nil
}

func parseRequestLineforTest(src []byte) (method Method, uri []byte, version []byte, next []byte, err error) {
	req := Request{}
	next, err = ParseRequestLine(&req, src)
	if err != nil {
		return MethodInvalid, nil, nil, nil, err
	}
	return req.Method, req.URI, req.Version, next, nil
}

var methodTable = [256]Method{}

var _ = func() int {
	//GET
	const GETIndex = 'G' ^ 'E' + 'T'
	methodTable[GETIndex] = MethodGET
	//PUT
	const PUTIndex = 'P' ^ 'U' + 'T'
	methodTable[PUTIndex] = MethodPUT
	//HEAD
	const HEADIndex = 'H' ^ 'E' + 'A'
	methodTable[HEADIndex] = MethodHEAD
	//POST
	const POSTIndex = 'P' ^ 'O' + 'S'
	methodTable[POSTIndex] = MethodPOST
	//BREW
	const BREWIndex = 'B' ^ 'R' + 'E'
	methodTable[BREWIndex] = MethodBREW
	//TRACE
	const TRACEIndex = 'T' ^ 'R' + 'A'
	methodTable[TRACEIndex] = MethodTRACE
	//PATCH
	const PATCHIndex = 'P' ^ 'A' + 'T'
	methodTable[PATCHIndex] = MethodPATCH
	//DELETE
	const DELETEIndex = 'D' ^ 'E' + 'L'
	methodTable[DELETEIndex] = MethodDELETE
	//CONNECT
	const CONNECTIndex = 'C' ^ 'O' + 'N'
	methodTable[CONNECTIndex] = MethodCONNECT
	//OPTIONS
	const OPTIONSIndex = 'O' ^ 'P' + 'T'
	methodTable[OPTIONSIndex] = MethodOPTIONS

	// all methods should have distinct index number
	var _ = map[int]Method{
		GETIndex:     MethodGET,
		PUTIndex:     MethodPUT,
		HEADIndex:    MethodHEAD,
		POSTIndex:    MethodPOST,
		BREWIndex:    MethodBREW,
		TRACEIndex:   MethodTRACE,
		PATCHIndex:   MethodPATCH,
		DELETEIndex:  MethodDELETE,
		CONNECTIndex: MethodCONNECT,
		OPTIONSIndex: MethodOPTIONS,
	}

	return 0
}()

func ParseRequestLine(dst *Request, src []byte) (next []byte, err error) {
	next = src
	var line []byte
	line, next, err = splitLine(next)
	if err != nil {
		return nil, err
	}
	MethodIndex := bytes.IndexByte(line, ' ')
	if MethodIndex < 0 || MethodIndex < 3 {
		return nil, ErrInvalidMethod
	}
	URIIndex := bytes.IndexByte(line[MethodIndex+1:], ' ')
	if URIIndex < 0 {
		return nil, ErrInvalidURI
	}
	dst.URI = line[MethodIndex+1 : MethodIndex+1+URIIndex]
	dst.Version = line[MethodIndex+1+URIIndex+1:]

	m := line[:MethodIndex]

	dst.Method = methodTable[m[0]^m[1]+m[2]]
	return next, nil
}

var ContentLengthHeader = []byte("Content-Length")

func ParseHeaders(dst *Request, src []byte) (next []byte, err error) {
	next = src
	var line []byte
	for {
		line, next, err = splitLine(next)
		if err != nil {
			return nil, err
		}
		if len(line) == 0 {
			break
		}
		if dst.lastHeader == nil {
			dst.Headers = GetHeader()
			dst.lastHeader = dst.Headers
		} else {
			dst.lastHeader.nextHeader = GetHeader()
			dst.lastHeader = dst.lastHeader.nextHeader
		}
		dst.lastHeader.raw = line
		dst.lastHeader.Name, dst.lastHeader.RawValue = ParseHeaderLine(line)

		if stricmp(dst.lastHeader.Name, ContentLengthHeader) {
			dst.ContentLength, err = ParseContentLength(dst.lastHeader.RawValue)
			if err != nil {
				return nil, err
			}
		}
	}
	return next, nil
}

func ParseContentLength(src []byte) (int64, error) {
	srcS := bytesToString(src)
	return strconv.ParseInt(srcS, 10, 64)
}

func ParseHeaderLine(src []byte) (name []byte, value []byte) {
	idx := bytes.IndexByte(src, ':')
	if idx < 0 {
		return src[:0], nil
	}
	// RFC2616 Section 4.2
	// Remove all leading and trailing LWS on field contents

	// skip leading LWS
	var i int = idx + 1
	for ; i < len(src); i++ {
		if src[i] != ' ' && src[i] != '\t' {
			break
		}
	}
	// skip trailing LWS
	var j int = len(src) - 1
	for ; j > i; j-- {
		if src[j] != ' ' && src[j] != '\t' {
			break
		}
	}
	return src[:idx], src[i : j+1]
}

type Header struct {
	noCopy
	raw []byte

	Name     []byte
	RawValue []byte

	nextHeader *Header // single linked list
}

var headerPool = sync.Pool{
	New: func() interface{} {
		return &Header{}
	},
}

func GetHeader() *Header {
	return headerPool.Get().(*Header)
}

func PutHeader(h *Header) {
	h.Reset()
	headerPool.Put(h)
}

func (h *Header) Reset() {
	h.raw = nil
	h.Name = nil
	h.RawValue = nil
	h.nextHeader = nil
}

func ReturnAllHeaders(h *Header) {
	for h != nil {
		next := h.nextHeader
		PutHeader(h)
		h = next
	}
}

const BufferPoolSize = 4096

var bufferPool = sync.Pool{
	New: func() interface{} {
		buffer := make([]byte, BufferPoolSize)
		return &buffer
	},
}

func GetBuffer() *[]byte {
	return bufferPool.Get().(*[]byte)
}

func PutBuffer(b *[]byte) {
	if cap(*b) >= BufferPoolSize {
		bufferPool.Put(b)
	}
}

var GlobalParserLock sync.Mutex

// Do not use this function in production code.
// This function is only for testing purpose.
// It is thread-safe but use global lock.
func ParseRequest(dst *Request, r io.Reader) (err error) {
	GlobalParserLock.Lock()
	defer GlobalParserLock.Unlock()
	dst.Reset()
	var buffer *[]byte = GetBuffer()
	//defer PutBuffer(buffer) // Allow GC to collect the buffer
	n, err := r.Read(*buffer)
	if err != nil {
		return err
	}
	var next []byte = (*buffer)[:n]
retryRead:

	// This function can't parse request line correctly if the request line is too long (>=4096)
	next, err = ParseRequestLine(dst, next)
	if err == ErrBufferTooSmall {
		buffer = GetBuffer()
		//defer PutBuffer(buffer)
		remainBytes := copy(*buffer, next)
		n, err = r.Read((*buffer)[remainBytes:])
		if err != nil {
			return err
		}
		next = (*buffer)[:remainBytes+n]
		goto retryRead
	} else if err != nil {
		return err
	}

	for {
		next, err = ParseHeaders(dst, next)
		if err == ErrBufferTooSmall {
			buffer = GetBuffer()
			//defer PutBuffer(buffer)
			remainBytes := copy(*buffer, next)
			n, err = r.Read((*buffer)[remainBytes:])
			if err != nil {
				return err
			}
			next = (*buffer)[:remainBytes+n]
			continue
		}
		if err != nil {
			return err
		}
		break
	}
	dst.next = next
	dst.Buffer = buffer
	return nil
}

func parseRequestForTest(data []byte) (*Request, error) {
	r := &Request{}
	err := ParseRequest(r, bytes.NewReader(data))
	return r, err
}

func parseRequestForTestIsValid(data []byte) bool {
	_, err := parseRequestForTest(data)
	return err == nil
}
