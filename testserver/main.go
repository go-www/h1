package main

import (
	"io"
	"log"
	"net"

	"github.com/go-www/h1"
	"github.com/kr/pretty"
)

type LogWriter struct {
	io.Writer
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	log.Println(string(p))
	return w.Writer.Write(p)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := h1.RequestReader{
		R:          conn,
		ReadBuffer: make([]byte, 8192),
		NextBuffer: nil,
		Request:    h1.Request{},
	}
	resp := h1.GetResponse(&LogWriter{conn})
	defer h1.PutResponse(resp)
	for {
		_, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				log.Println("EOF")
				return
			}
			log.Println(err)
			return
		}
		body, err := io.ReadAll(reader.Body())
		if err != nil {
			log.Println(err)
			return
		}

		data := []byte(pretty.Sprint(reader.Request))
		log.Println(string(data))
		resp.ContentLength = len(data) + len(body) + 2
		err = resp.WriteHeader(200)
		if err != nil {
			log.Println(err)
			return
		}
		resp.Write(body)
		resp.Write([]byte("\n\n"))
		resp.Write(data)
		if reader.Remaining() == 0 {
			err = resp.Flush()
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func main() {
	ln, err := net.Listen("tcp", ":50901")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Println(err)
		}
		go handleConnection(c)
	}
}
