package h1

import (
	"bytes"
	"errors"
)

type Request struct {
	noCopy

	method  Method
	URI     []byte
	Version []byte

	// Header
}

var (
	parserRequestMethodGET     = []byte("GET")
	parserRequestMethodPUT     = []byte("PUT")
	parserRequestMethodHEAD    = []byte("HEAD")
	parserRequestMethodPOST    = []byte("POST")
	parserRequestMethodTRACE   = []byte("TRACE")
	parserRequestMethodPATCH   = []byte("PATCH")
	parserRequestMethodDELETE  = []byte("DELETE")
	parserRequestMethodCONNECT = []byte("CONNECT")
	parserRequestMethodOPTIONS = []byte("OPTIONS")
)

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
	return req.method, req.URI, req.Version, next, nil
}

func ParseRequestLine(dst *Request, src []byte) (next []byte, err error) {
	next = src
	var line []byte
	line, next, err = splitLine(next)
	if err != nil {
		return nil, err
	}
	MethodIndex := bytes.IndexByte(line, ' ')
	if MethodIndex < 0 {
		return nil, ErrInvalidMethod
	}
	URIIndex := bytes.IndexByte(line[MethodIndex+1:], ' ')
	if URIIndex < 0 {
		return nil, ErrInvalidURI
	}
	dst.URI = line[MethodIndex+1 : MethodIndex+1+URIIndex]
	dst.Version = line[MethodIndex+1+URIIndex+1:]

	m := line[:MethodIndex]

	switch {
	case string(m) == string(parserRequestMethodGET):
		dst.method = MethodGET
	case string(m) == string(parserRequestMethodPUT):
		dst.method = MethodPUT
	case string(m) == string(parserRequestMethodHEAD):
		dst.method = MethodHEAD
	case string(m) == string(parserRequestMethodPOST):
		dst.method = MethodPOST
	case string(m) == string(parserRequestMethodTRACE):
		dst.method = MethodTRACE
	case string(m) == string(parserRequestMethodPATCH):
		dst.method = MethodPATCH
	case string(m) == string(parserRequestMethodDELETE):
		dst.method = MethodDELETE
	case string(m) == string(parserRequestMethodCONNECT):
		dst.method = MethodCONNECT
	case string(m) == string(parserRequestMethodOPTIONS):
		dst.method = MethodOPTIONS
	default:
		dst.method = MethodInvalid
		return nil, ErrInvalidMethod
	}
	return next, nil
}
