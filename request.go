package h1

import (
	"bytes"
	"errors"
	"strconv"
	"sync"
)

type Request struct {
	noCopy

	Method  Method
	URI     []byte
	Version []byte

	// Headers
	Headers    *Header
	lastHeader *Header

	ContentLength int64
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
	r.ContentLength = 0
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
	if idx == -1 {
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
	return strconv.ParseInt(string(src), 10, 64)
}

func ParseHeaderLine(src []byte) (name []byte, value []byte) {
	idx := bytes.IndexByte(src, ':')
	if idx < 0 {
		return src[:0], nil
	}
	// RFC2616 Section 4.2
	// Remove all leading and trailing LWS on field contents

	// skip leading LWS
	var i int = idx
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
