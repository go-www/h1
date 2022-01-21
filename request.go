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

	dst.method = methodTable[m[0]^m[1]+m[2]]
	return next, nil
}
