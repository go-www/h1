package h1

type Method uint8

const (
	MethodInvalid Method = iota
	MethodGET
	MethodHEAD
	MethodPOST
	MethodPUT
	MethodDELETE
	MethodCONNECT
	MethodOPTIONS
	MethodTRACE
	MethodPATCH
)

/*
	According to RFC7231 Section 4.1,
	All general purpose HTTP/1.1 servers MUST support the GET, HEAD.
*/