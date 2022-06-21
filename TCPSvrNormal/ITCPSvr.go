package TCPSvrNormal

import "net"

type ITcpNet interface {
	OnConnect(conn net.Conn) bool
	OnDisconnect(conn net.Conn)
	OnError(conn net.Conn, err error)
	OnReadData(conn net.Conn) (closed bool, err error)
}
