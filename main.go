package main

import (
	"io"
	"net"
)

type ReadType uint8

const (
	_ ReadType = iota
	ReadHead
	ReadContent
)

type test1 struct {
	readtype ReadType
}

func (t *test1) OnConnect(conn net.Conn) bool {
	println("连接上来了")
	return true
}
func (t *test1) OnDisconnect(conn net.Conn) {
	println("断开连接")
}
func (t *test1) OnError(conn net.Conn, err error) {
	println("连接错误")
}
func (t *test1) OnReadData(conn net.Conn) (closed bool, err error) {
	println("接收数据")
	err = nil
	b := make([]byte, 1)
	switch t.readtype {
	case ReadHead:
		{
			_, err = io.ReadFull(conn, b)
			t.readtype = ReadContent
		}
	case ReadContent:
		{
			_, err = io.ReadFull(conn, b)
			t.readtype = ReadHead
		}
	default:
		t.readtype = ReadHead
	}
	return false, err
}

func main() {
	tcpframe := NewNetFrame("127.0.0.1", 9998, &test1{})
	tcpframe.Start()
}
