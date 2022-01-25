package fuzz

import (
	"bytes"

	"github.com/go-www/h1"
)

func Fuzz(data []byte) int {
	var r h1.Request
	h1.ParseRequest(&r, bytes.NewReader(data))
	return 0
}
