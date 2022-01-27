package h1

type Connection uint8

const (
	ConnectionUnset Connection = iota
	ConnectionClose
	ConnectionKeepAlive
	ConnectionUpgrade
)

var connectionHeaderTable [16][]byte

const connectionHeaderTableMask = 1<<4 - 1

var _ = func() int {
	connectionHeaderTable[ConnectionClose] = []byte("Connection: close\r\n")
	connectionHeaderTable[ConnectionKeepAlive] = []byte("Connection: keep-alive\r\n")
	connectionHeaderTable[ConnectionUpgrade] = []byte("Connection: upgrade\r\n")

	return 0
}

func getConnectionHeader(c Connection) []byte {
	return connectionHeaderTable[c&connectionHeaderTableMask]
}
