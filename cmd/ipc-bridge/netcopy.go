package main

import (
	"io"
	"net"
)

var copyConn = netCopy

func netCopy(dst net.Conn, src net.Conn) (int64, error) {
	return io.Copy(dst, src)
}
