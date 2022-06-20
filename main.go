package main

import "net"

type test1 struct {
}

func (t test1) OnConnect(conn net.Conn) bool {
	println("连接上来了")
	return true
}
func (t test1) OnDisconnect(conn net.Conn) {
	println("断开连接")
}
func (t test1) OnError(conn net.Conn, err error) {
	println("连接错误")
}
func (t test1) OnRecv(conn net.Conn, Adata []byte, len int) bool {
	println("接收数据")
	return true
}

func main() {
	tcpframe := NewNetFrame("127.0.0.1", 9997, test1{})
	tcpframe.Start()
}
