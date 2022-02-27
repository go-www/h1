package h1

import (
	"bytes"
	"errors"

	"github.com/go-www/h1/encoding/percent"
)

type Query struct {
	Key   []byte
	Value []byte
}

type URI struct {
	RawURI []byte

	RawPath  []byte
	RawQuery []byte

	isQueryParsed bool
	queryArgs     []Query
}

func (u *URI) Reset() {
	u.RawURI = nil
	u.RawPath = nil
	u.RawQuery = nil
	u.isQueryParsed = false
	u.queryArgs = u.queryArgs[:0]
}

func (u *URI) Parse(uri []byte) {
	u.Reset()

	u.RawURI = uri

	// Parse the URI
	// Find the ?

	QIndex := bytes.IndexByte(uri, '?')
	u.RawPath = uri
	if QIndex != -1 {
		u.RawPath = uri[:QIndex]
		u.RawQuery = uri[QIndex+1:]
	}
}

func (u *URI) Path() []byte {
	return u.RawPath
}

func (u *URI) parseQuery() {
	next := u.RawQuery

	for {
		var key, value []byte

		// Find the Key
		keyEnd := bytes.IndexByte(next, '=')
		if keyEnd == -1 {
			break
		}

		key = next[:keyEnd]
		next = next[keyEnd+1:]

		// Find the Value

		valueEnd := bytes.IndexByte(next, '&')
		if valueEnd > 0 {
			value = next[:valueEnd]
			next = next[valueEnd+1:]
		} else {
			u.queryArgs = append(u.queryArgs, Query{
				Key:   key,
				Value: percent.Decode(next),
			})
			break
		}

		u.queryArgs = append(u.queryArgs, Query{
			Key:   key,
			Value: percent.Decode(value),
		})
	}
}

func (u *URI) Query() []Query {
	if !u.isQueryParsed {
		u.parseQuery()
		u.isQueryParsed = true
	}

	return u.queryArgs
}

var ErrKeyNotFound = errors.New("key not found")

func (u *URI) QueryValue(key []byte) ([]byte, error) {
	for _, q := range u.Query() {
		if bytes.Equal(q.Key, key) {
			return q.Value, nil
		}
	}

	return nil, ErrKeyNotFound
}
