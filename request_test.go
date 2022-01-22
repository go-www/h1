package h1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func Test_splitLine(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name     string
		args     args
		wantLine []byte
		wantRest []byte
		wantErr  bool
	}{
		{"empty", args{[]byte("")}, nil, nil, true},
		{"no newline", args{[]byte("hello")}, nil, []byte("hello"), true},
		{"newline", args{[]byte("hello\n")}, []byte("hello"), nil, false},
		{"crlf", args{[]byte("hello\r\n")}, []byte("hello"), nil, false},
		{"crlf2", args{[]byte("hello\r\nworld")}, []byte("hello"), []byte("world"), false},
		{"crlf3", args{[]byte("hello\r\nworld\r\n")}, []byte("hello"), []byte("world\r\n"), false},
		{"http", args{[]byte("POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 12\r\n\r\nHello World!")}, []byte("POST / HTTP/1.1"), []byte("Host: localhost\r\nContent-Length: 12\r\n\r\nHello World!"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotRest, err := splitLine(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("splitLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !bytes.Equal(gotLine, tt.wantLine) {
				t.Errorf("splitLine() gotLine = %v, want %v", gotLine, tt.wantLine)
			}
			if !bytes.Equal(gotRest, tt.wantRest) {
				t.Errorf("splitLine() gotRest = %v, want %v", gotRest, tt.wantRest)
			}
		})
	}
}

func Test_parseRequestLineforTest(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name        string
		args        args
		wantMethod  Method
		wantUri     []byte
		wantVersion []byte
		wantNext    []byte
		wantErr     bool
	}{
		{"empty", args{[]byte("")}, MethodInvalid, nil, nil, nil, true},
		{"no newline", args{[]byte("hello")}, MethodInvalid, nil, nil, nil, true},
		{"newline", args{[]byte("hello\n")}, MethodInvalid, nil, nil, nil, true},
		{"crlf", args{[]byte("hello\r\n")}, MethodInvalid, nil, nil, nil, true},
		{"HTTP1.1 GET", args{[]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodGET, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"HTTP1.1 HEAD", args{[]byte("HEAD / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodHEAD, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"HTTP1.1 POST", args{[]byte("POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 12\r\n\r\nHello World!")}, MethodPOST, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\nContent-Length: 12\r\n\r\nHello World!"), false},
		{"HTTP1.1 PUT", args{[]byte("PUT / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 12\r\n\r\nHello World!")}, MethodPUT, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\nContent-Length: 12\r\n\r\nHello World!"), false},
		{"HTTP1.1 DELETE", args{[]byte("DELETE /data?id=1 HTTP/1.1\r\nHost: localhost")}, MethodDELETE, []byte("/data?id=1"), []byte("HTTP/1.1"), []byte("Host: localhost"), false},
		{"HTTP1.1 CONNECT", args{[]byte("CONNECT / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodCONNECT, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"HTTP1.1 OPTIONS", args{[]byte("OPTIONS / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodOPTIONS, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"HTTP1.1 TRACE", args{[]byte("TRACE / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodTRACE, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"HTTP1.1 PATCH", args{[]byte("PATCH / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodPATCH, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"HTTP1.1 BREW", args{[]byte("BREW / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodBREW, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"Invalid Method 0", args{[]byte("INVALID / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodInvalid, []byte("/"), []byte("HTTP/1.1"), []byte("Host: localhost\r\n\r\n"), false},
		{"Invalid Method 1", args{[]byte("IN / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodInvalid, nil, nil, nil, true},
		{"Invalid URI", args{[]byte("GET HTTP/1.1\r\nHost: localhost\r\n\r\n")}, MethodInvalid, nil, nil, nil, true},
		{"Invalid Version", args{[]byte("GET /")}, MethodInvalid, nil, nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMethod, gotUri, gotVersion, gotNext, err := parseRequestLineforTest(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequestLineforTest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotMethod, tt.wantMethod) {
				t.Errorf("parseRequestLineforTest() gotMethod = %v, want %v", gotMethod, tt.wantMethod)
			}
			if !reflect.DeepEqual(gotUri, tt.wantUri) {
				t.Errorf("parseRequestLineforTest() gotUri = %v, want %v", gotUri, tt.wantUri)
			}
			if !reflect.DeepEqual(gotVersion, tt.wantVersion) {
				t.Errorf("parseRequestLineforTest() gotVersion = %v, want %v", gotVersion, tt.wantVersion)
			}
			if !reflect.DeepEqual(gotNext, tt.wantNext) {
				t.Errorf("parseRequestLineforTest() gotNext = %v, want %v", gotNext, tt.wantNext)
			}
		})
	}
}

func DeepCompare(a, b interface{}) bool {
	// check if the two interfaces are the same type
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	// check if the two interfaces are the pointer
	if reflect.ValueOf(a).Kind() == reflect.ValueOf(b).Kind() && reflect.ValueOf(a).Kind() == reflect.Ptr {
		return reflect.DeepEqual(reflect.ValueOf(a).Elem().Interface(), reflect.ValueOf(b).Elem().Interface())
	}

	// check if the two interfaces are the same value
	return reflect.DeepEqual(a, b)
}

func Test_parseRequestForTest(t *testing.T) {
	type args struct {
		data []byte
	}
	type TCase struct {
		name    string
		args    args
		want    *Request
		wantErr bool
	}
	tests := []TCase{
		{
			"empty",
			args{[]byte("")},
			&Request{},
			true,
		},
	}

	/*
		Header := &Header{
			raw:      []byte("Host: localhost\r\n\r\n"),
			Name:     []byte("Host"),
			RawValue: []byte("localhost"),
		}
			tests = append(tests, TCase{
				"get / request",
				args{[]byte(
					"GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
				)},
				&Request{
					Method:     MethodGET,
					URI:        []byte("/"),
					Version:    []byte("HTTP/1.1"),
					lastHeader: Header,
					Headers:    Header,
				},
				false,
			})
	*/
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRequestForTest(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRequestForTest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !DeepCompare(got, tt.want) {
				t.Errorf("parseRequestForTest() = %v, want %v", got, tt.want)
				gotDumped, _ := json.MarshalIndent(got, "", "  ")
				wantDumped, _ := json.MarshalIndent(tt.want, "", "  ")
				t.Logf("got: %s\n", gotDumped)
				t.Logf("want: %s\n", wantDumped)
			}
		})
	}
}

func Test_parseRequestForTestIsValid(t *testing.T) {
	type args struct {
		data []byte
	}

	var LargeHeaderRequest bytes.Buffer
	LargeHeaderRequest.WriteString("GET / HTTP/1.1\r\nHost: localhost\r\n")
	for i := 0; i < 1024; i++ {
		LargeHeaderRequest.WriteString(fmt.Sprintf("Header%d: value%d\r\n", i, i))
	}
	LargeHeaderRequest.WriteString("\r\n")

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{[]byte("")}, false},
		{"GET / request", args{[]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, true},
		{"POST / request", args{[]byte("POST / HTTP/1.1\r\nHost: localhost\r\nContent-Length: 12\r\n\r\nHello World!")}, true},
		{"Malformed GET Request0", args{[]byte("get / HTTP/1.1\r\nHost: localhost\r\n\r\n")}, true},
		{"Malformed GET Request1", args{[]byte("GET /HTTP/1.1\r\nHost: localhost\r\n\r\n")}, false},
		{"Malformed GET Request2", args{[]byte("GET/ HTTP/1.1\r\nHost: localhost\r\n\r\n")}, false},
		{"Large Request Line", args{LargeHeaderRequest.Bytes()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRequestForTestIsValid(tt.args.data); got != tt.want {
				t.Errorf("parseRequestForTestIsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
